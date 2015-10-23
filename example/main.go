// Package main provides an example of the stampzilla/gocast library
package main

import (
	"fmt"
	"time"

	"github.com/stampzilla/gocast/discovery"
	"github.com/stampzilla/gocast/events"
)

func main() {
	discovery := discovery.NewService()

	go discoveryListner(discovery)

	// Start a periodic discovery
	fmt.Println("Start discovery")
	discovery.Periodic(time.Second * 10)
	<-time.After(time.Second * 30)

	fmt.Println("Stop discovery")
	discovery.Stop()

	select {}
}

func discoveryListner(discovery *discovery.Service) {
	for device := range discovery.Found() {
		fmt.Printf("New device discoverd: %#v \n", device)

		//plexHandler := NewPlexHandler()
		//device.Subscribe("urn:x-cast:plex", plexHandler)
		//device.Subscribe("urn:x-cast:com.google.cast.media", mediaHandler)

		device.OnEvent(func(event events.Event) {
			switch data := event.(type) {
			case events.Connected:
				fmt.Println(device.Name(), "- Connected, weeihoo")
			case events.Disconnected:
				fmt.Println(device.Name(), "- Disconnected, bah :/")
			case events.AppStarted:
				fmt.Println(device.Name(), "- App started:", data.DisplayName, "(", data.AppID, ")")
			case events.AppStopped:
				fmt.Println(device.Name(), "- App stopped:", data.DisplayName, "(", data.AppID, ")")
			//gocast.MediaEvent:
			//plexEvent:
			default:
				fmt.Printf("unexpected event %T: %#v\n", data, data)
			}
		})

		device.Connect()
	}
}
