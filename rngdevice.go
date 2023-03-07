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

const (
	RngDeviceType = "rngdevice"
	RngDevRandom  = "/dev/random"
	RngDevUrandom = "/dev/urandom"
)

// RngDevice represents a random number generator device.
type RngDevice struct {
	// DeviceType string `default:"rngdevice" yaml:"device-type"`

	// ID is the device ID
	ID string `yaml:"id"`

	// Driver is the device driver
	Driver DeviceDriver `yaml:"driver"`

	// Bus is the bus path name of a this device.
	Bus string `yaml:"bus"`

	// Addr is the address offset of this device on the bus.
	Addr string `yaml:"address"`

	// Filename is entropy source on the host
	Filename string `yaml:"filename"`

	// MaxBytes is the bytes allowed to guest to get from the host’s entropy per period
	MaxBytes uint `yaml:"max-bytes"`

	// Period is duration of a read period in seconds
	Period uint `yaml:"period"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"ccw-dev-no"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`
}

// Valid returns true if the RngDevice structure is valid and complete.
func (r RngDevice) Valid() error {
	if r.ID == "" {
		return fmt.Errorf("RngDevice has empty ID field")
	}

	if r.Driver == "" {
		return fmt.Errorf("RngDevice has empty Driver field")
	}

	return nil
}

// QemuParams returns the qemu parameters built out of the RngDevice.
func (r RngDevice) QemuParams(config *Config) []string {
	var qemuParams []string

	//-object rng-random,filename=/dev/hwrng,id=rng0
	var objectParams []string
	//-device virtio-rng-pci,rng=rng0,max-bytes=1024,period=1000
	var deviceParams []string

	objectParams = append(objectParams, "rng-random")
	objectParams = append(objectParams, "id="+r.ID)

	deviceParams = append(deviceParams, r.deviceName(config))
	deviceParams = append(deviceParams, "rng="+r.ID)

	if r.Bus != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", r.Bus))
	}

	// virtio can have a BusAddr since they are pci devices
	addr := config.pciBusSlots.GetSlot(r.Addr)
	if addr > 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("addr=0x%02x", addr))
	}

	if r.Transport.isVirtioPCI(config) && r.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", r.ROMFile))
	}

	if r.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, "iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", r.DevNo))
	}

	if r.Filename != "" {
		objectParams = append(objectParams, "filename="+r.Filename)
	}

	if r.MaxBytes > 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("max-bytes=%d", r.MaxBytes))
	}

	if r.Period > 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("period=%d", r.Period))
	}

	qemuParams = append(qemuParams, "-object")
	qemuParams = append(qemuParams, strings.Join(objectParams, ","))

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (r RngDevice) deviceName(config *Config) string {
	if r.Transport == "" {
		r.Transport = r.Transport.defaultTransport(config)
	}

	if r.Driver != VirtioRng {
		return string(r.Driver)
	}

	// handle VirtioRng
	switch r.Transport {
	case TransportPCI:
		return "virtio-rng-pci"
	case TransportCCW:
		return "virtio-rng-ccw"
	case TransportMMIO:
		return "virtio-rng-device"
	default:
		return ""
	}
}
