package gocast

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/handlers"
)

func (d *Device) listener() {
	for {
		packet := d.wrapper.Read()

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

		catched := false
		for _, subscription := range d.subscriptions {
			if subscription.Receive(message, &headers) {
				catched = true
			}
		}

		if !catched {
			fmt.Println("LOST MESSAGE:")
			spew.Dump(message)
		}
	}
}

func (d *Device) Connect() error {
	event := ConnectedEvent{}
	d.Dispatch(event)

	//log.Printf("connecting to %s:%d ...", d.ip, d.port)

	var err error
	d.conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", d.ip, d.port), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return fmt.Errorf("Failed to connect to Chromecast. Error:%s", err)
	}

	d.wrapper = NewPacketStream(d.conn)
	go d.listener()

	d.Subscribe("urn:x-cast:com.google.cast.tp.connection", d.connectionHandler)
	d.Subscribe("urn:x-cast:com.google.cast.tp.heartbeat", d.heartbeatHandler)
	d.Subscribe("urn:x-cast:com.google.cast.receiver", d.receiverHandler)

	return nil
}

func (d *Device) Send(urn, sourceId, destinationId string, payload handlers.Headers) error {
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

	//spew.Dump("Writing", message)

	_, err = d.wrapper.Write(&data)

	return err
}
