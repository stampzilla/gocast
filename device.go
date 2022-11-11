package gocast

import (
	"context"
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

	stop context.CancelFunc

	eventListners []func(event events.Event)
	subscriptions map[string]*Subscription

	connectionHandler Handler
	heartbeatHandler  *handlers.Heartbeat
	ReceiverHandler   *handlers.Receiver

	logger logrus.FieldLogger
}

func NewDevice(logger logrus.FieldLogger) *Device {
	d := &Device{
		eventListners:     make([]func(event events.Event), 0),
		reconnect:         make(chan struct{}),
		subscriptions:     make(map[string]*Subscription),
		connectionHandler: &handlers.Connection{},
		heartbeatHandler:  handlers.NewHeartbeat(),
		ReceiverHandler:   &handlers.Receiver{},
		logger:            logger,
	}

	return d
}
func (d *Device) SetLogger(logger logrus.FieldLogger) {
	d.Lock()
	d.logger = logger
	d.Unlock()
}
func (d *Device) SetName(name string) {
	d.Lock()
	d.name = name
	d.Unlock()
}

func (d *Device) SetUuid(uuid string) {
	d.Lock()
	d.uuid = uuid
	d.Unlock()
}

func (d *Device) SetIp(ip net.IP) {
	d.Lock()
	d.ip = ip
	d.Unlock()
}

func (d *Device) SetPort(port int) {
	d.Lock()
	d.port = port
	d.Unlock()
}

func (d *Device) Name() string {
	d.RLock()
	defer d.RUnlock()
	return d.name
}

func (d *Device) Uuid() string {
	d.RLock()
	defer d.RUnlock()
	return d.uuid
}

func (d *Device) Ip() net.IP {
	d.RLock()
	defer d.RUnlock()
	return d.ip
}

func (d *Device) Port() int {
	d.RLock()
	defer d.RUnlock()
	return d.port
}

func (d *Device) getConn() net.Conn {
	d.RLock()
	defer d.RUnlock()
	return d.conn
}
func (d *Device) getSubscriptionsAsSlice() []*Subscription {
	d.RLock()
	subs := make([]*Subscription, len(d.subscriptions))
	i := 0
	for _, v := range d.subscriptions {
		subs[i] = v
		i++
	}
	defer d.RUnlock()
	return subs
}

func (d *Device) String() string {
	return d.Name() + " - " + d.ip.String() + ":" + strconv.Itoa(d.port)
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

	d.logger.Debug("Subscribing to ", urn, " --- ", destinationId)
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
	for _, sub := range subs {
		delete(d.subscriptions, sub)
	}
	d.Unlock()
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
