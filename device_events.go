package gocast

import "github.com/stampzilla/gocast/events"

func (d *Device) OnEvent(callback func(event events.Event)) {
	d.eventListners = append(d.eventListners, callback)
}

func (d *Device) Dispatch(event events.Event) {
	for _, callback := range d.eventListners {
		go callback(event)
	}
}
