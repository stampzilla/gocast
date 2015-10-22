package gocast

import (
	"net"
	"strconv"
)

type Device struct {
	name string
	ip   net.IP
	port int
}

func NewDevice() *Device {
	return &Device{}
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

func (d *Device) Name() string {
	return d.name
}

func (d *Device) String() string {
	return d.name + " - " + d.ip.String() + ":" + strconv.Itoa(d.port)
}
