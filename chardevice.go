/*
// Copyright contributors to the Virtual Machine Manager for Go project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/

// Package qemu provides methods and types for launching and managing QEMU
// instances.  Instances can be launched with the LaunchQemu function and
// managed thereafter via QMPStart and the QMP object that this function
// returns.  To manage a qemu instance after it has been launched you need
// to pass the -qmp option during launch requesting the qemu instance to create
// a QMP unix domain manageent socket, e.g.,
// -qmp unix:/tmp/qmp-socket,server,nowait.  For more information see the
// example below.

package qcli

import (
	"fmt"
	"strings"
)

// CharDeviceBackend is the character device backend for qemu
type CharDeviceBackend string

const (
	// Pipe creates a 2 way connection to the guest.
	Pipe CharDeviceBackend = "pipe"

	// Socket creates a 2 way stream socket (TCP or Unix).
	Socket CharDeviceBackend = "socket"

	// CharConsole sends traffic from the guest to QEMU's standard output.
	CharConsole CharDeviceBackend = "console"

	// Serial sends traffic from the guest to a serial device on the host.
	Serial CharDeviceBackend = "serial"

	// TTY is an alias for Serial.
	TTY CharDeviceBackend = "tty"

	// PTY creates a new pseudo-terminal on the host and connect to it.
	PTY CharDeviceBackend = "pty"

	// File sends traffic from the guest to a file on the host.
	File CharDeviceBackend = "file"

	// Stdio creates a 2 way connection to guest via stdio
	Stdio CharDeviceBackend = "stdio"

	// SpiceVMC creates a spice-protocol char device over a virtserialport
	SpiceVMC CharDeviceBackend = "spicevmc"
)

// CharDevice represents a qemu character device.
type CharDevice struct {
	Backend CharDeviceBackend `yaml:"backend"`

	// Driver is the qemu device driver
	Driver DeviceDriver `yaml:"driver"`

	// Bus is the serial bus associated to this device.
	Bus string `yaml:"bus"`

	// DeviceID is the user defined device ID.
	DeviceID string `yaml:"device-id"`

	ID   string `yaml:"id"`
	Path string `yaml:"path"`
	Name string `yaml:"name"`

	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool `yaml:"disable-modern"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"ccw-dev-no"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`

	// Mux will multiplex output if value is 'on', 'off' disables, default value
	Mux string `yaml:"multiplex"`

	// Signal will enable signal processing if 'on', or not if 'off'
	Signal string `yaml:"signal"`
}

// VirtioSerialTransport is a map of the virtio-serial device name that
// corresponds to each transport.
var VirtioSerialTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-serial-pci",
	TransportCCW:  "virtio-serial-ccw",
	TransportMMIO: "virtio-serial-device",
}

// Valid returns nil if the CharDevice structure is valid and complete.
func (cdev CharDevice) Valid() error {
	if cdev.ID == "" {
		return fmt.Errorf("CharDevice missing ID value: %+v", cdev)
	}
	// Stdio backend does not require a path
	if cdev.Backend != Stdio && cdev.Path == "" {
		return fmt.Errorf("CharDevice with Backend='%s' must have Path", cdev.Backend)
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this character device.
func (cdev CharDevice) QemuParams(config *Config) []string {
	var cdevParams []string
	var deviceParams []string
	var qemuParams []string

	deviceParams = append(deviceParams, cdev.deviceName(config))
	if cdev.Driver == VirtioSerial {
		if s := cdev.Transport.disableModern(config, cdev.DisableModern); s != "" {
			deviceParams = append(deviceParams, s)
		}
	}
	if cdev.Bus != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", cdev.Bus))
	}
	deviceParams = append(deviceParams, fmt.Sprintf("chardev=%s", cdev.ID))
	deviceParams = append(deviceParams, fmt.Sprintf("id=%s", cdev.DeviceID))
	if cdev.Name != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("name=%s", cdev.Name))
	}
	if cdev.Driver == VirtioSerial && cdev.Transport.isVirtioPCI(config) && cdev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", cdev.ROMFile))
	}

	if cdev.Driver == VirtioSerial && cdev.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, "iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", cdev.DevNo))
	}

	cdevParams = append(cdevParams, string(cdev.Backend))
	cdevParams = append(cdevParams, fmt.Sprintf("id=%s", cdev.ID))
	switch cdev.Backend {
	case Socket:
		cdevParams = append(cdevParams, fmt.Sprintf("path=%s,server=on,wait=off", cdev.Path))
	case File:
		cdevParams = append(cdevParams, fmt.Sprintf("path=%s", cdev.Path))
	}

	cParam := getConfigOnOff("Mux", "mux", cdev.Mux)
	if cParam != "" {
		cdevParams = append(cdevParams, cParam)
	}

	cParam = getConfigOnOff("Signal", "signal", cdev.Signal)
	if cParam != "" {
		cdevParams = append(cdevParams, cParam)
	}

	// Legacy serial is special. It does not follow the device + driver model
	if cdev.Driver != LegacySerial && cdev.Driver != PCISerialDevice {
		qemuParams = append(qemuParams, "-device")
		qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	}

	//appending -serial none and -monitor none to qemuparams if we are using pci-serial
	if cdev.Driver == PCISerialDevice {
		qemuParams = append(qemuParams, "-serial")
		qemuParams = append(qemuParams, "none")
	}

	qemuParams = append(qemuParams, "-chardev")
	qemuParams = append(qemuParams, strings.Join(cdevParams, ","))

	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (cdev CharDevice) deviceName(config *Config) string {
	if cdev.Transport == "" {
		cdev.Transport = cdev.Transport.defaultTransport(config)
	}

	switch cdev.Driver {
	case VirtioSerial:
		return VirtioSerialTransport[cdev.Transport]
	}

	return string(cdev.Driver)
}
