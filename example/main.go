// Package main provides an example of the stampzilla/gocast library
package main

import (
	"fmt"
	"time"

	"github.com/stampzilla/gocast"
	"github.com/stampzilla/gocast/discovery"
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
		fmt.Println(device)

		//plexHandler := NewPlexHandler()
		//device.Subscribe("urn:x-cast:plex", plexHandler)
		//device.Subscribe("urn:x-cast:com.google.cast.media", mediaHandler)

		device.OnEvent(func(event gocast.Event) {
			switch data := event.(type) {
			case gocast.ConnectedEvent:
				fmt.Println("Connected, weeihoo")
			case gocast.DisconnectedEvent:
				fmt.Println("Disconnected, bah :/")
			//gocast.RecevierEvent:
			//gocast.MediaEvent:
			//plexEvent:
			default:
				fmt.Printf("unexpected event %T: %#v\n", data, data)
			}
		})

		device.Connect()
	}
}
