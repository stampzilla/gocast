package gocast

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/stampzilla/gocast/handlers"
)

type Device struct {
	name string
	ip   net.IP
	port int
	conn net.Conn

	eventListners []func(event Event)

	connectionHandler *handlers.Connection
	heartbeatHandler  *handlers.Heartbeat
	receiverHandler   *handlers.Receiver
}

func NewDevice() *Device {
	return &Device{
		eventListners: make([]func(event Event), 0),

		connectionHandler: &handlers.Connection{},
		heartbeatHandler:  &handlers.Heartbeat{},
		receiverHandler:   &handlers.Receiver{},
	}
}

func (d *Device) SetName(name string) {
	d.name = name
}

func (d *Device) SetIp(ip net.IP) {
	d.ip = ip
}

func (d *Device) SetPort(port int) {
	d.port = port
}

func (d *Device) Subscribe(urn string, handler Handler) {
	handler.SendCallback(d.SendCallback)
	handler.Connect()
}

func (d *Device) OnEvent(callback func(event Event)) {
	d.eventListners = append(d.eventListners, callback)
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) String() string {
	return d.name + " - " + d.ip.String() + ":" + strconv.Itoa(d.port)
}

func (d *Device) SendCallback(headers handlers.Headers) {
	fmt.Printf("Oh send!: %#v\n", headers)
}

func (d *Device) Dispatch(event Event) {
	for _, callback := range d.eventListners {
		go callback(event)
	}
}

func (d *Device) Connect() error {
	event := ConnectedEvent{}
	d.Dispatch(event)

	log.Printf("connecting to %s:%d ...", d.ip, d.port)

	var err error
	d.conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", d.ip, d.port), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return fmt.Errorf("Failed to connect to Chromecast. Error:%s", err)
	}

	d.Subscribe("urn:x-cast:com.google.cast.tp.connection", d.connectionHandler)
	d.Subscribe("urn:x-cast:com.google.cast.tp.heartbeat", d.heartbeatHandler)
	d.Subscribe("urn:x-cast:com.google.cast.receiver", d.receiverHandler)

	return nil
}
