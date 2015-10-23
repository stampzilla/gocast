package gocast

func (d *Device) OnEvent(callback func(event Event)) {
	d.eventListners = append(d.eventListners, callback)
}

func (d *Device) Dispatch(event Event) {
	for _, callback := range d.eventListners {
		go callback(event)
	}
}
