// Package discovery provides a discovery service for chromecast devices
package discovery

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/zeroconf/v2"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast"
)

const queryTimeout = 3 * time.Second

type entry struct {
	ip   net.IP
	port int
	info map[string]string
}

type Service struct {
	found     chan *gocast.Device
	entriesCh chan entry

	foundDevices    map[string]*gocast.Device
	periodicRunning uint32
	stop            context.CancelFunc
}

func NewService() *Service {
	return &Service{
		found:        make(chan *gocast.Device),
		entriesCh:    make(chan entry),
		foundDevices: make(map[string]*gocast.Device),
	}
}

func (d *Service) periodic(ctx context.Context, interval time.Duration) error {
	if i := atomic.LoadUint32(&d.periodicRunning); i != 0 {
		return fmt.Errorf("Periodic discovery is already running")
	}

	d.query(ctx)

	ticker := time.NewTicker(interval)
	atomic.AddUint32(&d.periodicRunning, 1)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d.query(ctx)
			case <-ctx.Done():
				logrus.Debug("stopping periodic goroutine")
				return
			}
		}
	}()

	return nil
}

// query browses for chromecasts for a few seconds and forwards every complete
// entry to the listener.
//
// zeroconf sends its queries from port 5353, so devices answer with multicast.
// That is required for devices on other subnets reached through an mDNS
// reflector, which only repeats multicast and never forwards the unicast
// replies that queries from a random source port get.
func (d *Service) query(ctx context.Context) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	results := make(chan *zeroconf.ServiceEntry)
	browseDone := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			select {
			case e, ok := <-results:
				if !ok {
					return
				}
				if len(e.AddrIPv4) == 0 {
					continue
				}
				select {
				case d.entriesCh <- entry{
					ip:   e.AddrIPv4[0],
					port: e.Port,
					info: decodeTxtRecord(e.Text),
				}:
				case <-ctx.Done():
				}
			case <-browseDone:
				// Browse has returned and sends nothing more. It does not close
				// results if it failed before it started listening.
				return
			}
		}
	}()

	// Browse blocks until c is done.
	err := zeroconf.Browse(c, "_googlecast._tcp", "local.", results, zeroconf.SelectIPTraffic(zeroconf.IPv4))
	if err != nil {
		logrus.Error("error doing mdns query: ", err)
	}
	close(browseDone)
	<-readerDone
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
			info := entry.info
			key := info["id"] // Use device ID as key, allowes the device to change IP
			if key == "" {
				continue
			}

			if dev, ok := d.foundDevices[key]; ok {
				changed := false
				if !entry.ip.Equal(dev.Ip()) {
					dev.SetIp(entry.ip)
					changed = true
				}
				if entry.port != dev.Port() {
					dev.SetPort(entry.port)
					changed = true
				}
				if changed {
					dev.SetLogger(logrus.WithFields(logrus.Fields{
						"ip":   entry.ip.String(),
						"port": entry.port,
						"name": info["fn"],
					}))
				}
				continue
			}

			device := gocast.NewDevice(logrus.WithFields(logrus.Fields{
				"ip":   entry.ip.String(),
				"port": entry.port,
				"name": info["fn"],
			}))

			device.SetIp(entry.ip)
			device.SetPort(entry.port)

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
				// Nobody received it, forget it so it is offered again on the next scan.
				logrus.Warnf("discovery: no receiver for device %s, will retry on next scan", info["fn"])
				delete(d.foundDevices, key)
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

func decodeTxtRecord(txt []string) map[string]string {
	m := make(map[string]string)

	for _, v := range txt {
		k, val, ok := strings.Cut(v, "=")
		if !ok {
			continue
		}
		if k == "fn" { // Friendly name
			val = decodeDnsEntry(val)
		}
		m[k] = val
	}

	return m
}
