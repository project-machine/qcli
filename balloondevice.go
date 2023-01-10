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

package qemu

import (
	"fmt"
	"strings"
)

// BalloonDevice represents a memory balloon device.
type BalloonDevice struct {
	DeflateOnOOM  bool   `yaml:"deflate-on-oom"`
	DisableModern bool   `yaml:"disable-modern"`
	ID            string `yaml:"id"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"ccw-dev-no"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`
}

// BalloonDeviceTransport is a map of the virtio-balloon device name that
// corresponds to each transport.
var BalloonDeviceTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-balloon-pci",
	TransportCCW:  "virtio-balloon-ccw",
	TransportMMIO: "virtio-balloon-device",
}

// QemuParams returns the qemu parameters built out of the BalloonDevice.
func (b BalloonDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	deviceParams = append(deviceParams, b.deviceName(config))

	if b.ID != "" {
		deviceParams = append(deviceParams, "id="+b.ID)
	}

	if b.Transport.isVirtioPCI(config) && b.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", b.ROMFile))
	}

	if b.Transport.isVirtioCCW(config) {
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", b.DevNo))
	}

	if b.DeflateOnOOM {
		deviceParams = append(deviceParams, "deflate-on-oom=on")
	} else {
		deviceParams = append(deviceParams, "deflate-on-oom=off")
	}
	if s := b.Transport.disableModern(config, b.DisableModern); s != "" {
		deviceParams = append(deviceParams, s)
	}
	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// Valid returns true if the balloonDevice structure is valid and complete.
func (b BalloonDevice) Valid() error {
	if b.ID == "" {
		return fmt.Errorf("Invalid BalloonDevice, ID field is unset")
	}
	return nil
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (b BalloonDevice) deviceName(config *Config) string {
	if b.Transport == "" {
		b.Transport = b.Transport.defaultTransport(config)
	}

	return BalloonDeviceTransport[b.Transport]
}
