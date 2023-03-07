package qcli

import (
	"fmt"
	"strings"
)

const RemoteDisplayPortBase = 5900
const SpiceSerialNamespace = "com.redhat.spice.0"
const SpiceCharDevDriver = "spicevmc"
const SpiceCharDevName = "vdagent"

// SpiceDevice represents a qemu spice protocol device.
type SpiceDevice struct {
	ID               string `yaml:"id"`
	Port             string `yaml:"port"`
	HostAddress      string `yaml:"host-address"`
	TLSPort          string `yaml:"tls-port"`
	DisableTicketing bool   `yaml:"disable-ticketing"`
	// FIXME: implement the rest of -spice
}

// Valid returns true if there is a valid structure defined for SpiceDevice
func (dev SpiceDevice) Valid() error {
	if dev.Port == "" && dev.TLSPort == "" {
		return fmt.Errorf("SpiceDevice 'Port' or 'TLSPort' value is required")
	}

	if dev.Port != "" && dev.TLSPort != "" {
		return fmt.Errorf("SpiceDevice has 'Port' and 'TLSPort' set, only one allowed")
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this spice device.
func (dev SpiceDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string
	var virtportParams []string
	var chardevParams []string

	if dev.Port != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("port=%s", dev.Port))
	}
	if dev.TLSPort != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("tls-port=%s", dev.TLSPort))
	}

	addr := "127.0.0.1"
	if dev.HostAddress != "" {
		addr = dev.HostAddress
	}
	deviceParams = append(deviceParams, fmt.Sprintf("addr=%s", addr))

	if dev.DisableTicketing {
		deviceParams = append(deviceParams, fmt.Sprintf("disable-ticketing=on"))
	}

	// add the virtserialport to enable copy-paste if guest is configured
	//  -device virtserialport,chardev=spicechannel0,name=com.redhat.spice.0
	chardevID := "spicechannel0"
	virtportParams = append(virtportParams, "virtserialport")
	virtportParams = append(virtportParams, fmt.Sprintf("chardev=%s", chardevID))
	virtportParams = append(virtportParams, fmt.Sprintf("name=%s", SpiceSerialNamespace))

	//  -chardev spicevmc,id=spicechannel0,name=vdagent
	chardevParams = append(chardevParams, SpiceCharDevDriver)
	chardevParams = append(chardevParams, fmt.Sprintf("id=%s", chardevID))
	chardevParams = append(chardevParams, fmt.Sprintf("name=%s", SpiceCharDevName))

	qemuParams = append(qemuParams, "-spice")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	qemuParams = append(qemuParams, "-device", "virtio-serial-pci")
	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(virtportParams, ","))
	qemuParams = append(qemuParams, "-chardev")
	qemuParams = append(qemuParams, strings.Join(chardevParams, ","))

	return qemuParams
}
