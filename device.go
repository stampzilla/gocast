package gocast

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/handlers"
)

type Device struct {
	name string
	ip   net.IP
	port int
	conn net.Conn
	id   int

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
	sourceId := "sender-0"
	destinationId := "receiver-0"

	callback := func(payload handlers.Headers) error {

		return d.SendCallback(urn, sourceId, destinationId, payload)
	}

	handler.SendCallback(callback)
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

func (d *Device) SendCallback(urn, sourceId, destinationId string, payload handlers.Headers) error {
	fmt.Printf("Oh send!: %#v\n", payload)

	d.id++
	payload.RequestId = &d.id

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to json.Marshal: ", err)
		return err
	}
	payloadString := string(payloadJson)

	message := &api.CastMessage{
		ProtocolVersion: api.CastMessage_CASTV2_1_0.Enum(),
		SourceId:        &sourceId,
		DestinationId:   &destinationId,
		Namespace:       &urn,
		PayloadType:     api.CastMessage_STRING.Enum(),
		PayloadUtf8:     &payloadString,
	}

	proto.SetDefaults(message)

	data, err := proto.Marshal(message)
	if err != nil {
		fmt.Println("Failed to proto.Marshal: ", err)
		return err
	}

	spew.Dump("Writing", message)

	_, err = d.conn.Write(data)

	return err
}

func (d *Device) listener() {
	wrapper := NewPacketStream(d.conn)

	for {
		packet := wrapper.Read()

		message := &api.CastMessage{}
		err := proto.Unmarshal(*packet, message)
		if err != nil {
			log.Fatalf("Failed to unmarshal CastMessage: %s", err)
		}

		//spew.Dump("Message!", message)

		var headers handlers.Headers

		err = json.Unmarshal([]byte(*message.PayloadUtf8), &headers)

		if err != nil {
			log.Fatalf("Failed to unmarshal message: %s", err)
		}

		//for _, channel := range client.channels {
		//channel.message(message, &headers)
		//}

		log.Printf("RECEIVED: %#v\n", message)

	}
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

	go d.listener()

	<-time.After(time.Second)

	d.Subscribe("urn:x-cast:com.google.cast.tp.connection", d.connectionHandler)
	d.Subscribe("urn:x-cast:com.google.cast.tp.heartbeat", d.heartbeatHandler)
	d.Subscribe("urn:x-cast:com.google.cast.receiver", d.receiverHandler)

	return nil
}
