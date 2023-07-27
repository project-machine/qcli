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

// LegacySerialDevice represents a qemu legacy serial device.
type LegacySerialDevice struct {
	// specify a chardev-id of an existing CharDev, and use the name
	ChardevID string `yaml:"chardev-id"`
	Name      string `yaml:"name"`
	MonMux    bool   `yaml:"mon-mux-enable"`
	// Set if needing to multiplex serial and HMP monitor output togeter on stdio
	Backend CharDeviceBackend `yaml:"backend"`
	Path    string            `yaml:"path"`
}

// Valid returns true if the LegacySerialDevice structure is valid and complete.
func (dev LegacySerialDevice) Valid() error {
	if dev.MonMux {
		return nil
	}
	if dev.Backend == "" {
		// One must be set
		if dev.Name == "" && dev.ChardevID == "" {
			return fmt.Errorf("LegacySerialDevice requires either Name or ChardevID field to be set")
		}

		// Name and ChardevID are mutually exclusive
		if dev.Name != "" && dev.ChardevID != "" {
			return fmt.Errorf("LegacySerialDevice Name and ChardevID field are mutually exclusive")
		}
	} else {
		if dev.Backend != Socket {
			return fmt.Errorf("LegacySerialDevice only supports Backend='unix'")
		}
		if dev.Path == "" {
			return fmt.Errorf("LegacySerialDevice with Backend must have Path")
		}
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this serial device.
func (dev LegacySerialDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var sdevParams []string

	if dev.MonMux {
		sdevParams = append(sdevParams, "mon:stdio")
	} else {
		if dev.Backend == Socket {
			sdevParams = append(sdevParams, fmt.Sprintf("unix:%s,server=on,wait=off", dev.Path))
		} else {
			if dev.Name != "" && dev.ChardevID == "" {
				sdevParams = append(sdevParams, dev.Name)
			}
			if dev.ChardevID != "" && dev.Name == "" {
				sdevParams = append(sdevParams, fmt.Sprintf("chardev:%s", dev.ChardevID))
			}
		}
	}

	qemuParams = append(qemuParams, "-serial")
	qemuParams = append(qemuParams, strings.Join(sdevParams, ","))

	return qemuParams
}

/* Not used currently
// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (dev LegacySerialDevice) deviceName(config *Config) string {
	return dev.Chardev
}
*/

// SerialDevice represents a qemu serial device.
type SerialDevice struct {
	// Driver is the qemu device driver
	Driver DeviceDriver

	// ID is the serial device identifier.
	ID string

	// PCI Slot
	Addr string

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string

	// MaxPorts is the maximum number of ports for this device. (note: 1, 2, or 4 for pci-serial driver)
	MaxPorts uint

	//Enable Multifunction
	Multifunction bool

	//virtio-serial specific attributes
	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string

	// Transport is the virtio transport for this device.
	Transport VirtioTransport

	//pci-serial specific attributes
	//Chardev associated with PCISerialDevice
	ChardevIDs []string
}

// Valid returns true if the SerialDevice structure is valid and complete.
func (dev SerialDevice) Valid() error {
	if dev.Driver == "" {
		return fmt.Errorf("SerialDevice has empty Driver field")
	}
	if dev.ID == "" {
		return fmt.Errorf("SerialDevice has empty ID field")
	}
	if dev.Driver == PCISerialDevice {
		if len(dev.ChardevIDs) > 4 || len(dev.ChardevIDs) == 0 {
			return fmt.Errorf("PCISerialDeviceDevice has a malformed list of ChardevIDs (length 0 or length > 4)")
		}
		if dev.ChardevIDs[0] == "" {
			return fmt.Errorf("PCISerialDeviceDevice has no associated ChardevID")
		}
		if dev.MaxPorts != 1 && dev.MaxPorts != 2 && dev.MaxPorts != 4 {
			return fmt.Errorf("PCISerialDeviceDevice has MaxPorts not equal to 1, 2, or 4")
		}
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this serial device.
func (dev SerialDevice) QemuParams(config *Config) []string {
	var deviceParams []string
	var qemuParams []string

	deviceParams = append(deviceParams, dev.deviceName(config))
	deviceParams = append(deviceParams, fmt.Sprintf("id=%s", dev.ID))
	if dev.Addr != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("addr=%s", dev.Addr))
	}
	if dev.ROMFile != "" {
		if dev.Driver == PCISerialDevice || dev.Transport.isVirtioPCI(config) {
			deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", dev.ROMFile))
		}
	}
	if dev.Multifunction {
		deviceParams = append(deviceParams, "multifunction=on")
	}
	switch dev.Driver {
	case VirtioSerial:
		if s := dev.Transport.disableModern(config, dev.DisableModern); s != "" {
			deviceParams = append(deviceParams, s)
		}
		if dev.Transport.isVirtioPCI(config) && dev.MaxPorts != 0 {
			deviceParams = append(deviceParams, fmt.Sprintf("max_ports=%d", dev.MaxPorts))
		}
		if dev.Transport.isVirtioCCW(config) {
			if config.Knobs.IOMMUPlatform {
				deviceParams = append(deviceParams, "iommu_platform=on")
			}
			deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", dev.DevNo))
		}
	case PCISerialDevice:
		if dev.MaxPorts == 1 {
			deviceParams = append(deviceParams, fmt.Sprintf("chardev=%s", dev.ChardevIDs[0]))
		} else {
			deviceParams = append(deviceParams, fmt.Sprintf("chardev1=%s", dev.ChardevIDs[0]))
			if len(dev.ChardevIDs) > 1 && dev.ChardevIDs[1] != "" {
				deviceParams = append(deviceParams, fmt.Sprintf("chardev2=%s", dev.ChardevIDs[1]))
			}
			if dev.MaxPorts == 4 {
				if len(dev.ChardevIDs) > 2 && dev.ChardevIDs[2] != "" {
					deviceParams = append(deviceParams, fmt.Sprintf("chardev3=%s", dev.ChardevIDs[2]))
				}
				if len(dev.ChardevIDs) == 4 && dev.ChardevIDs[3] != "" {
					deviceParams = append(deviceParams, fmt.Sprintf("chardev4=%s", dev.ChardevIDs[3]))
				}
			}
		}
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (dev SerialDevice) deviceName(config *Config) string {
	var devNameStr string
	if dev.Driver == VirtioSerial {
		if dev.Transport == "" {
			dev.Transport = dev.Transport.defaultTransport(config)
		}
		devNameStr = VirtioSerialTransport[dev.Transport]
	}
	if dev.Driver == PCISerialDevice {
		add_str := ""
		if dev.MaxPorts == 2 {
			add_str = "-2x"
		} else if dev.MaxPorts == 4 {
			add_str = "-4x"
		}
		devNameStr = fmt.Sprintf("%s%s", dev.Driver, add_str)
	}
	return devNameStr
}
