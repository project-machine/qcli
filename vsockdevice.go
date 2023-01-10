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
	"os"
	"strings"
)

// VSOCKDevice represents a AF_VSOCK socket.
type VSOCKDevice struct {
	ID string

	ContextID uint64

	// VHostFD vhost file descriptor that holds the ContextID
	VHostFD *os.File

	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string

	// Transport is the virtio transport for this device.
	Transport VirtioTransport
}

// VSOCKDeviceTransport is a map of the vhost-vsock device name that
// corresponds to each transport.
var VSOCKDeviceTransport = map[VirtioTransport]string{
	TransportPCI:  "vhost-vsock-pci",
	TransportCCW:  "vhost-vsock-ccw",
	TransportMMIO: "vhost-vsock-device",
}

const (
	// MinimalGuestCID is the smallest valid context ID for a guest.
	MinimalGuestCID uint64 = 3

	// MaxGuestCID is the largest valid context ID for a guest.
	MaxGuestCID uint64 = 1<<32 - 1
)

const (
	// VSOCKGuestCID is the VSOCK guest CID parameter.
	VSOCKGuestCID = "guest-cid"
)

// Valid returns true if the VSOCKDevice structure is valid and complete.
func (vsock VSOCKDevice) Valid() error {
	if vsock.ID == "" {
		return fmt.Errorf("VSOCKDevicve has empty ID field")
	}
	if vsock.ContextID < MinimalGuestCID {
		return fmt.Errorf("VSOCKDevicve has ContextID < MinimalCID (%d < %d) fields", vsock.ContextID, MinimalGuestCID)
	}
	if vsock.ContextID > MaxGuestCID {
		return fmt.Errorf("VSOCKDevicve has ContextID > MaxGuestCID (%d > %d) fields", vsock.ContextID, MaxGuestCID)
	}

	return nil
}

// QemuParams returns the qemu parameters built out of the VSOCK device.
func (vsock VSOCKDevice) QemuParams(config *Config) []string {
	var deviceParams []string
	var qemuParams []string

	driver := vsock.deviceName(config)
	deviceParams = append(deviceParams, driver)
	if s := vsock.Transport.disableModern(config, vsock.DisableModern); s != "" {
		deviceParams = append(deviceParams, s)
	}
	if vsock.VHostFD != nil {
		qemuFDs := config.appendFDs([]*os.File{vsock.VHostFD})
		deviceParams = append(deviceParams, fmt.Sprintf("vhostfd=%d", qemuFDs[0]))
	}
	deviceParams = append(deviceParams, fmt.Sprintf("id=%s", vsock.ID))
	deviceParams = append(deviceParams, fmt.Sprintf("%s=%d", VSOCKGuestCID, vsock.ContextID))

	if vsock.Transport.isVirtioPCI(config) && vsock.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", vsock.ROMFile))
	}

	if vsock.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, "iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", vsock.DevNo))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (vsock VSOCKDevice) deviceName(config *Config) string {
	if vsock.Transport == "" {
		vsock.Transport = vsock.Transport.defaultTransport(config)
	}

	return VSOCKDeviceTransport[vsock.Transport]
}
