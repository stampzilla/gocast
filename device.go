package gocast

import (
	"net"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/handlers"
)

type Device struct {
	sync.RWMutex
	name      string
	uuid      string
	ip        net.IP
	port      int
	connected bool
	conn      net.Conn
	wrapper   *packetStream
	reconnect chan struct{}

	eventListners []func(event events.Event)
	subscriptions map[string]*Subscription

	connectionHandler Handler
	heartbeatHandler  *handlers.Heartbeat
	ReceiverHandler   *handlers.Receiver
}

func NewDevice() *Device {
	d := &Device{
		eventListners:     make([]func(event events.Event), 0),
		reconnect:         make(chan struct{}),
		subscriptions:     make(map[string]*Subscription),
		connectionHandler: &handlers.Connection{},
		heartbeatHandler:  &handlers.Heartbeat{},
		ReceiverHandler:   &handlers.Receiver{},
	}

	d.heartbeatHandler.OnFailure = func() {
		d.Disconnect()
	}

	return d
}

func (d *Device) SetName(name string) {
	d.name = name
}

func (d *Device) SetUuid(uuid string) {
	d.uuid = uuid
}

func (d *Device) SetIp(ip net.IP) {
	d.ip = ip
}

func (d *Device) SetPort(port int) {
	d.port = port
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) Uuid() string {
	return d.uuid
}

func (d *Device) Ip() net.IP {
	return d.ip
}

func (d *Device) Port() int {
	return d.port
}

func (d *Device) String() string {
	return d.name + " - " + d.ip.String() + ":" + strconv.Itoa(d.port)
}

func (d *Device) Subscribe(urn, destinationId string, handler Handler) {
	sourceId := "sender-0"

	s := &Subscription{
		Urn:           urn,
		SourceId:      sourceId,
		DestinationId: destinationId,
		Handler:       handler,
		Device:        d,
		inFlight:      make(map[int]chan *api.CastMessage),
	}

	d.Lock()
	d.subscriptions[s.Sha256()] = s
	d.Unlock()

	handler.RegisterSend(s.Send)
	handler.RegisterRequest(s.Request)
	handler.RegisterDispatch(d.Dispatch)
	handler.Connect()

	logrus.Debug("Subscribing to ", urn, " --- ", destinationId)
}

func (d *Device) UnsubscribeByUrn(urn string) {
	subs := []string{}
	d.RLock()
	for k, s := range d.subscriptions {
		if s.Urn == urn {
			subs = append(subs, k)
		}
	}
	d.RUnlock()
	d.Lock()
	defer d.Unlock()
	for _, sub := range subs {
		delete(d.subscriptions, sub)
	}
}

func (d *Device) UnsubscribeByUrnAndDestinationId(urn, destinationId string) {
	subs := []string{}
	d.RLock()
	for k, s := range d.subscriptions {
		if s.Urn == urn && s.DestinationId == destinationId {
			s.Handler.Disconnect()
			subs = append(subs, k)
		}
	}
	d.RUnlock()
	d.Lock()
	defer d.Unlock()
	for _, sub := range subs {
		delete(d.subscriptions, sub)
	}
}
