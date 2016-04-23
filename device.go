package gocast

import (
	"net"
	"strconv"

	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/handlers"
)

type Device struct {
	name      string
	uuid      string
	ip        net.IP
	port      int
	conn      net.Conn
	wrapper   *packetStream
	reconnect chan struct{}

	eventListners []func(event events.Event)
	subscriptions []*Subscription

	connectionHandler Handler
	heartbeatHandler  Handler
	ReceiverHandler   *handlers.Receiver
}

func NewDevice() *Device {
	return &Device{
		eventListners:     make([]func(event events.Event), 0),
		reconnect:         make(chan struct{}),
		connectionHandler: &handlers.Connection{},
		heartbeatHandler:  &handlers.Heartbeat{},
		ReceiverHandler:   &handlers.Receiver{},
	}
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
	//destinationId := "receiver-0"

	s := &Subscription{
		Urn:           urn,
		SourceId:      sourceId,
		DestinationId: destinationId,
		Handler:       handler,
		Device:        d,
		inFlight:      make(map[int]chan *api.CastMessage),
	}

	//log.Println("Subscribing to ", urn, " --- ", destinationId)

	//callback := func(payload handlers.Headers) error {
	//return d.Send(urn, sourceId, destinationId, payload)
	//}

	d.subscriptions = append(d.subscriptions, s)

	handler.RegisterSend(s.Send)
	handler.RegisterRequest(s.Request)
	handler.RegisterDispatch(d.Dispatch)
	handler.Connect()
}
