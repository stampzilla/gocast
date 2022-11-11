// Package discovery provides a discovery service for chromecast devices
package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/micro/mdns"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast"
)

type Service struct {
	found     chan *gocast.Device
	entriesCh chan *mdns.ServiceEntry

	foundDevices    map[string]*gocast.Device
	periodicRunning uint32
	stop            context.CancelFunc
}

func NewService() *Service {
	return &Service{
		found:        make(chan *gocast.Device),
		entriesCh:    make(chan *mdns.ServiceEntry),
		foundDevices: make(map[string]*gocast.Device),
	}
}

func (d *Service) periodic(ctx context.Context, interval time.Duration) error {
	if i := atomic.LoadUint32(&d.periodicRunning); i != 0 {
		return fmt.Errorf("Periodic discovery is already running")
	}

	c, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err := mdns.Query(&mdns.QueryParam{
		Service: "_googlecast._tcp",
		Domain:  "local",
		Timeout: time.Second * 1,
		Entries: d.entriesCh,
		Context: c,
	})

	if err != nil {
		logrus.Error("error doing mdns query: ", err)
	}

	ticker := time.NewTicker(interval)
	atomic.AddUint32(&d.periodicRunning, 1)
	go func() {
		for {
			err := mdns.Query(&mdns.QueryParam{
				Service: "_googlecast._tcp",
				Domain:  "local",
				Timeout: time.Second * 1,
				Entries: d.entriesCh,
				Context: c,
			})
			if err != nil {
				logrus.Error("error doing mdns query: ", err)
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				ticker.Stop()
				logrus.Debug("stopping periodic goroutine")
				return
			}
		}
	}()

	return nil
}

func (d *Service) Start(pCtx context.Context, interval time.Duration) {
	ctx, cancel := context.WithCancel(pCtx)
	d.stop = cancel

	go d.listner(ctx)
	err := d.periodic(ctx, interval)
	if err != nil {
		logrus.Error("error starting periodic mdns query: ", err)
	}
}

func (d *Service) Stop() {
	if d.stop != nil {
		d.stop()
	}
}

func (d *Service) Found() chan *gocast.Device {
	return d.found
}

func (d *Service) listner(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logrus.Debug("stopping listner goroutine")
			d.foundDevices = make(map[string]*gocast.Device)
			return
		case entry := <-d.entriesCh:
			// fmt.Printf("Got new entry: %#v\n", entry)

			name := strings.Split(entry.Name, "._googlecast")

			// Skip everything that dont have googlecast in the fdqn
			if len(name) < 2 {
				continue
			}

			info := decodeTxtRecord(entry.Info)
			key := info["id"] // Use device ID as key, allowes the device to change IP

			if dev, ok := d.foundDevices[key]; ok {
				changed := false
				if !entry.AddrV4.Equal(dev.Ip()) {
					dev.SetIp(entry.AddrV4)
					changed = true
				}
				if entry.Port != dev.Port() {
					dev.SetPort(entry.Port)
					changed = true
				}
				if changed {
					dev.SetLogger(logrus.WithFields(logrus.Fields{
						"ip":   entry.AddrV4.String(),
						"port": entry.Port,
						"name": info["fn"],
					}))
				}
				continue
			}

			device := gocast.NewDevice(logrus.WithFields(logrus.Fields{
				"ip":   entry.AddrV4.String(),
				"port": entry.Port,
				"name": info["fn"],
			}))

			device.SetIp(entry.AddrV4)
			device.SetPort(entry.Port)

			device.SetUuid(key)
			device.SetName(info["fn"])

			d.foundDevices[key] = device

			delay := time.NewTimer(time.Second)
			select {
			case d.found <- device:
				if !delay.Stop() {
					<-delay.C
				}
			case <-delay.C:
			}
		}
	}
}

func decodeDnsEntry(text string) string {
	text = strings.ReplaceAll(text, `\.`, ".")
	text = strings.ReplaceAll(text, `\ `, " ")

	re := regexp.MustCompile(`([\\][0-9][0-9][0-9])`)
	text = re.ReplaceAllStringFunc(text, func(source string) string {
		i, err := strconv.Atoi(source[1:])
		if err != nil {
			return ""
		}

		return string([]byte{byte(i)})
	})

	return text
}

func decodeTxtRecord(txt string) map[string]string {
	m := make(map[string]string)

	s := strings.Split(txt, "|")
	for _, v := range s {
		s := strings.Split(v, "=")
		if len(s) == 2 {
			if s[0] == "fn" { // Friendly name
				m[s[0]] = decodeDnsEntry(s[1])
			} else {
				m[s[0]] = s[1]
			}
		}
	}

	return m
}
