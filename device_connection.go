package gocast

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

func (d *Device) reader(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil {
				logrus.Errorf("closing reader %s: %s", d.Name(), ctx.Err())
			}
			return
		case p := <-d.wrapper.packets:
			if p.err != nil {
				logrus.Errorf("Error reading from chromecast error: %s Packet: %#v", p.err, p)
				return
			}
			packet := p.payload

			message := &api.CastMessage{}
			err := proto.Unmarshal(packet, message)
			if err != nil {
				logrus.Errorf("Failed to unmarshal CastMessage: %s", err)
				continue
			}

			headers := &responses.Headers{}

			err = json.Unmarshal([]byte(*message.PayloadUtf8), headers)

			if err != nil {
				logrus.Errorf("Failed to unmarshal message: %s", err)
				continue
			}

			catched := false
			for _, subscription := range d.getSubscriptionsAsSlice() {
				if subscription.Receive(message, headers) {
					catched = true
				}
			}

			if !catched {
				logrus.Debug("LOST MESSAGE:")
				logrus.Debug(spew.Sdump(message))
			}
		}
	}
}

func (d *Device) Connect(ctx context.Context) error {
	d.heartbeatHandler.OnFailure = func() { // make sure we reconnect if we loose heartbeat
		logrus.Errorf("heartbeat timeout for: %s trying to reconnect", d.Name())

		d.Disconnect()
		for { // try to connect until no error
			err := d.connect(ctx)
			if err == nil {
				break
			}
			logrus.Error("error reconnect: ", err)
			time.Sleep(2 * time.Second)
		}
	}
	return d.connect(ctx)
}

func (d *Device) connect(pCtx context.Context) error {
	ctx, cancel := context.WithCancel(pCtx)
	d.stop = cancel

	ip := d.Ip()
	port := d.Port()

	logrus.Infof("connecting to %s:%d ...", ip, port)

	if d.getConn() != nil {
		err := d.conn.Close()
		if err != nil {
			logrus.Error("trying to connect with existing connection. error closing: ", err)
		}
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to Chromecast. Error:%s", err)
	}

	d.Lock()
	d.conn = conn
	d.connected = true
	d.Unlock()

	d.wrapper = NewPacketStream(d.conn)
	go d.wrapper.readPackets(ctx)
	go d.reader(ctx)

	d.Subscribe("urn:x-cast:com.google.cast.tp.connection", "receiver-0", d.connectionHandler)
	d.Subscribe("urn:x-cast:com.google.cast.tp.heartbeat", "receiver-0", d.heartbeatHandler)
	d.Subscribe("urn:x-cast:com.google.cast.receiver", "receiver-0", d.ReceiverHandler)

	d.Dispatch(events.Connected{})

	return nil
}

func (d *Device) Disconnect() {
	logrus.Debug("disconnecting: ", d.Name())

	for _, subscription := range d.getSubscriptionsAsSlice() {
		logrus.Debugf("disconnect subscription %s: %s ", d.Name(), subscription.Urn)
		subscription.Handler.Disconnect()
	}
	d.Lock()
	d.subscriptions = make(map[string]*Subscription)
	d.Unlock()

	if d.stop != nil { // make sure any old goroutines are stopped
		d.stop()
	}

	if c := d.getConn(); d != nil {
		c.Close()
		d.Lock()
		d.conn = nil
		d.Unlock()
	}

	d.Lock()
	d.connected = false
	d.Unlock()

	d.Dispatch(events.Disconnected{})
}

func (d *Device) Send(urn, sourceId, destinationId string, payload responses.Payload) error {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
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
		return err
	}

	if *message.Namespace != "urn:x-cast:com.google.cast.tp.heartbeat" {
		logrus.Debugf("Writing to %s: %s", d.Name(), spew.Sdump(message))
	}

	if d.conn == nil {
		return fmt.Errorf("we are disconnected, cannot send!")
	}

	_, err = d.wrapper.Write(data)

	return err
}
