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

// VFIODevice represents a qemu vfio device meant for direct access by guest OS.
type VFIODevice struct {
	// Bus-Device-Function of device
	BDF string

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string

	// VendorID specifies vendor id
	VendorID string

	// DeviceID specifies device id
	DeviceID string

	// Bus specifies device bus
	Bus string

	// Transport is the virtio transport for this device.
	Transport VirtioTransport
}

// VFIODeviceTransport is a map of the vfio device name that corresponds to
// each transport.
var VFIODeviceTransport = map[VirtioTransport]string{
	TransportPCI:  "vfio-pci",
	TransportCCW:  "vfio-ccw",
	TransportMMIO: "vfio-device",
}

// Valid returns true if the VFIODevice structure is valid and complete.
func (vfioDev VFIODevice) Valid() error {
	if vfioDev.BDF == "" {
		return fmt.Errorf("VFIODevice has empty BDF field")
	}
	return nil
}

// QemuParams returns the qemu parameters built out of this vfio device.
func (vfioDev VFIODevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	driver := vfioDev.deviceName(config)

	deviceParams = append(deviceParams, fmt.Sprintf("%s,host=%s", driver, vfioDev.BDF))
	if vfioDev.Transport.isVirtioPCI(config) {
		if vfioDev.VendorID != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("x-pci-vendor-id=%s", vfioDev.VendorID))
		}
		if vfioDev.DeviceID != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("x-pci-device-id=%s", vfioDev.DeviceID))
		}
		if vfioDev.ROMFile != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", vfioDev.ROMFile))
		}
	}

	if vfioDev.Bus != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", vfioDev.Bus))
	}

	if vfioDev.Transport.isVirtioCCW(config) {
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", vfioDev.DevNo))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (vfioDev VFIODevice) deviceName(config *Config) string {
	if vfioDev.Transport == "" {
		vfioDev.Transport = vfioDev.Transport.defaultTransport(config)
	}

	return VFIODeviceTransport[vfioDev.Transport]
}
