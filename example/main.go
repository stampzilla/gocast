// Package main provides an example of the stampzilla/gocast library
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/discovery"
	"github.com/stampzilla/gocast/events"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	discovery := discovery.NewService()

	go discoveryListner(discovery)

	// Start a periodic discovery
	fmt.Println("Start discovery")
	discovery.Start(context.Background(), time.Second*10)
	<-time.After(time.Second * 15)

	fmt.Println("Stop discovery")
	discovery.Stop()

	select {}
}

func discoveryListner(discovery *discovery.Service) {
	for device := range discovery.Found() {
		fmt.Printf("New device discovered: %#v \n", device)

		// plexHandler := NewPlexHandler()
		// device.Subscribe("urn:x-cast:plex", plexHandler)
		// device.Subscribe("urn:x-cast:com.google.cast.media", mediaHandler)

		d := device
		device.OnEvent(func(event events.Event) {
			switch data := event.(type) {
			case events.Connected:
				fmt.Println(d.Name(), "- Connected, weeihoo")
			case events.Disconnected:
				fmt.Println(d.Name(), "- Disconnected, bah :/")
			case events.AppStarted:
				fmt.Println(d.Name(), "- App started:", data.DisplayName, "(", data.AppID, ")")
			case events.AppStopped:
				fmt.Println(d.Name(), "- App stopped:", data.DisplayName, "(", data.AppID, ")")
			// gocast.MediaEvent:
			// plexEvent:
			default:
				fmt.Printf("unexpected event %T: %#v\n", data, data)
			}
		})

		device.Connect(context.Background())

		//go func() {
		//<-time.After(time.Second * 10)
		//device.Disconnect()
		//}()
	}
}
