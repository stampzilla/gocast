// Package discovery provides a discovery service for chromecast devices
package discovery

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/micro/mdns"
	"github.com/stampzilla/gocast"
)

type Service struct {
	found     chan *gocast.Device
	entriesCh chan *mdns.ServiceEntry

	foundDevices map[string]*gocast.Device
	stopPeriodic chan struct{}
}

func NewService() *Service {
	s := &Service{
		found:        make(chan *gocast.Device),
		entriesCh:    make(chan *mdns.ServiceEntry),
		foundDevices: make(map[string]*gocast.Device, 0),
	}

	go s.listner()

	return s
}

func (d *Service) Periodic(interval time.Duration) error {
	if d.stopPeriodic != nil {
		return fmt.Errorf("Periodic discovery is already running")
	}

	mdns.Query(&mdns.QueryParam{
		Service: "_googlecast._tcp",
		Domain:  "local",
		Timeout: time.Second * 1,
		Entries: d.entriesCh,
	})

	ticker := time.NewTicker(interval)
	d.stopPeriodic = make(chan struct{})
	go func() {
		for {
			mdns.Query(&mdns.QueryParam{
				Service: "_googlecast._tcp",
				Domain:  "local",
				Timeout: time.Second * 1,
				Entries: d.entriesCh,
			})
			select {
			case <-ticker.C:
			case <-d.stopPeriodic:
				ticker.Stop()
				d.foundDevices = make(map[string]*gocast.Device, 0)

				return
			}
		}
	}()

	return nil
}

func (d *Service) Stop() {
	if d.stopPeriodic != nil {
		close(d.stopPeriodic)
		d.stopPeriodic = nil
	}
}

func (d *Service) Found() chan *gocast.Device {
	return d.found
}

func (d *Service) listner() {
	for entry := range d.entriesCh {
		// fmt.Printf("Got new entry: %#v\n", entry)

		name := strings.Split(entry.Name, "._googlecast")

		// Skip everything that dont have googlecast in the fdqn
		if len(name) < 2 {
			continue
		}

		info := decodeTxtRecord(entry.Info)
		key := info["id"] // Use device ID as key, allowes the device to change IP

		if dev, ok := d.foundDevices[key]; ok {
			// If not connected, update address and reconnect
			if !dev.Connected() {
				dev.SetIp(entry.AddrV4)
				dev.SetPort(entry.Port)
				dev.Reconnect()
			}
			// Skip already connected devices
			continue
		}

		device := gocast.NewDevice()
		device.SetIp(entry.AddrV4)
		device.SetPort(entry.Port)

		device.SetUuid(key)
		device.SetName(info["fn"])

		d.foundDevices[key] = device

		select {
		case d.found <- device:
		case <-time.After(time.Second):
		}
	}
}

func decodeDnsEntry(text string) string {
	text = strings.Replace(text, `\.`, ".", -1)
	text = strings.Replace(text, `\ `, " ", -1)

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
