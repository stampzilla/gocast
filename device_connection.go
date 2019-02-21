package gocast

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

func (d *Device) reader() {
	for {
		packet, err := d.wrapper.Read()

		if err != nil {
			log.Println("Error reading from chromecast error:", err, "Packet:", packet)
			d.Disconnect()
			//d.reconnect <- struct{}{}
			return
		}

		message := &api.CastMessage{}
		err = proto.Unmarshal(*packet, message)
		if err != nil {
			log.Fatalf("Failed to unmarshal CastMessage: %s", err)
			continue
		}

		//spew.Dump("Message!", message)

		headers := &responses.Headers{}

		err = json.Unmarshal([]byte(*message.PayloadUtf8), headers)

		if err != nil {
			log.Fatalf("Failed to unmarshal message: %s", err)
			continue
		}

		catched := false
		d.RLock()
		for _, subscription := range d.subscriptions {
			if subscription.Receive(message, headers) {
				catched = true
			}
		}
		d.RUnlock()

		if !catched {
			log.Println("LOST MESSAGE:")
			spew.Dump(message)
		}
	}
}

func (d *Device) Connected() bool {
	d.RLock()
	defer d.RUnlock()
	return d.connected
}
func (d *Device) Connect() error {
	go d.reconnector()
	return d.connect()
}
func (d *Device) Reconnect() {
	select {
	case d.reconnect <- struct{}{}:
	default:
	}
}
func (d *Device) reconnector() {
	for {
		select {
		case <-d.reconnect:
			log.Println("Reconnect signal received")
			time.Sleep(time.Second * 2)
			d.connect()
		}
	}
}
func (d *Device) connect() error {
	log.Printf("connecting to %s:%d ...", d.ip, d.port)

	if d.conn != nil {
		return fmt.Errorf("Already connected to: %s (%s:%d)", d.Name(), d.Ip().String(), d.Port())
	}

	var err error
	d.conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", d.ip, d.port), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		//d.reconnect <- struct{}{}
		return fmt.Errorf("Failed to connect to Chromecast. Error:%s", err)
	}

	d.Lock()
	d.connected = true
	d.Unlock()

	event := events.Connected{}
	d.Dispatch(event)

	d.wrapper = NewPacketStream(d.conn)
	go d.reader()

	d.Subscribe("urn:x-cast:com.google.cast.tp.connection", "receiver-0", d.connectionHandler)
	d.Subscribe("urn:x-cast:com.google.cast.tp.heartbeat", "receiver-0", d.heartbeatHandler)
	d.Subscribe("urn:x-cast:com.google.cast.receiver", "receiver-0", d.ReceiverHandler)

	return nil
}

func (d *Device) Disconnect() {
	d.Lock()
	if d.conn != nil {
		for _, subscription := range d.subscriptions {
			subscription.Handler.Disconnect()
		}

		d.subscriptions = make(map[string]*Subscription, 0)
		d.Dispatch(events.Disconnected{})

		d.conn.Close()
		d.conn = nil
	}

	d.connected = false
	d.Unlock()
}

func (d *Device) Send(urn, sourceId, destinationId string, payload responses.Payload) error {
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

	if *message.Namespace != "urn:x-cast:com.google.cast.tp.heartbeat" {
		log.Println("Writing:", spew.Sdump(message))
	}

	if d.conn == nil {
		return fmt.Errorf("We are disconnected, cannot send!")
	}

	_, err = d.wrapper.Write(data)

	return err
}
