package gocast

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

func (d *Device) reader(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil {
				d.logger.Errorf("closing reader:  %s", ctx.Err())
			}
			return
		case p := <-d.wrapper.packets:
			if p.err != nil {
				d.logger.Errorf("Error reading from chromecast error: %s Packet: %#v", p.err, p)
				return
			}
			packet := p.payload

			message := &api.CastMessage{}
			err := proto.Unmarshal(packet, message)
			if err != nil {
				d.logger.Errorf("Failed to unmarshal CastMessage: %s", err)
				continue
			}

			headers := &responses.Headers{}

			err = json.Unmarshal([]byte(*message.PayloadUtf8), headers)

			if err != nil {
				d.logger.Errorf("Failed to unmarshal message: %s", err)
				continue
			}

			catched := false
			for _, subscription := range d.getSubscriptionsAsSlice() {
				if subscription.Receive(message, headers) {
					catched = true
				}
			}

			if !catched {
				d.logger.Debug("LOST MESSAGE:")
				d.logger.Debug(spew.Sdump(message))
			}
		}
	}
}

func (d *Device) Connect(ctx context.Context) error {
	d.heartbeatHandler.OnFailure = func() { // make sure we reconnect if we loose heartbeat
		d.logger.Errorf("heartbeat timeout, trying to reconnect")

		d.Disconnect()
		for { // try to connect until no error
			err := d.connect(ctx)
			if err == nil {
				break
			}
			d.logger.Error("error reconnect: ", err)
			time.Sleep(2 * time.Second)
		}
	}
	return d.connect(ctx)
}

func (d *Device) connect(pCtx context.Context) error {
	ctx, cancel := context.WithCancel(pCtx)
	d.stop = cancel

	d.logger.Info("connecting")

	if d.getConn() != nil {
		err := d.conn.Close()
		if err != nil {
			d.logger.Error("trying to connect with existing connection. error closing: ", err)
		}
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", d.Ip(), d.Port()), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to Chromecast. Error:%s", err)
	}

	d.Lock()
	d.conn = conn
	d.Unlock()

	d.wrapper = NewPacketStream(d.conn, d.logger)
	go d.wrapper.readPackets(ctx)
	go d.reader(ctx)

	d.Subscribe("urn:x-cast:com.google.cast.tp.connection", "receiver-0", d.connectionHandler)
	d.Subscribe("urn:x-cast:com.google.cast.tp.heartbeat", "receiver-0", d.heartbeatHandler)
	d.Subscribe("urn:x-cast:com.google.cast.receiver", "receiver-0", d.ReceiverHandler)

	d.Dispatch(events.Connected{})

	return nil
}

func (d *Device) Disconnect() {
	d.logger.Debug("disconnecting: ", d.Name())

	for _, subscription := range d.getSubscriptionsAsSlice() {
		d.logger.Debugf("disconnect subscription: %s ", subscription.Urn)
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
		d.logger.Debugf("Writing: %s", spew.Sdump(message))
	}

	if d.conn == nil {
		return fmt.Errorf("we are disconnected, cannot send!")
	}

	_, err = d.wrapper.Write(data)

	return err
}
