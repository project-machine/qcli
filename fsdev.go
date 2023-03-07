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

// Virtio9PMultidev filesystem behaviour to deal
// with multiple devices being shared with a 9p export.
type Virtio9PMultidev string

const (
	// Remap shares multiple devices with only one export.
	Remap Virtio9PMultidev = "remap"

	// Warn assumes that only one device is shared by the same export.
	// Only a warning message is logged (once) by qemu on host side.
	// This is the default behaviour.
	Warn Virtio9PMultidev = "warn"

	// Forbid like "warn" but also deny access to additional devices on guest.
	Forbid Virtio9PMultidev = "forbid"
)

// FSDriver represents a qemu filesystem driver.
type FSDriver string

// SecurityModelType is a qemu filesystem security model type.
type SecurityModelType string

const (
	// Local is the local qemu filesystem driver.
	Local FSDriver = "local"

	// Handle is the handle qemu filesystem driver.
	Handle FSDriver = "handle"

	// Proxy is the proxy qemu filesystem driver.
	Proxy FSDriver = "proxy"
)

const (
	// None is like passthrough without failure reports.
	None SecurityModelType = "none"

	// PassThrough uses the same credentials on both the host and guest.
	PassThrough SecurityModelType = "passthrough"

	// MappedXattr stores some files attributes as extended attributes.
	MappedXattr SecurityModelType = "mapped-xattr"

	// MappedFile stores some files attributes in the .virtfs directory.
	MappedFile SecurityModelType = "mapped-file"
)

// FSDevice represents a qemu filesystem configuration.
type FSDevice struct {
	// Driver is the qemu device driver
	Driver DeviceDriver `yaml:"driver"`

	// FSDriver is the filesystem driver backend.
	FSDriver FSDriver `yaml:"fs-driver"`

	// ID is the filesystem identifier.
	ID string `yaml:"id"`

	// Path is the host root path for this filesystem.
	Path string `yaml:"path"`

	// MountTag is the device filesystem mount point tag.
	MountTag string `yaml:"mount-tag"`

	// SecurityModel is the security model for this filesystem device.
	SecurityModel SecurityModelType `yaml:"security-model"`

	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool `yaml:"disable-modern"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"ccw-dev-no"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`

	// Multidev is the filesystem behaviour to deal
	// with multiple devices being shared with a 9p export
	Multidev Virtio9PMultidev `yaml:"multidev"`
}

// Virtio9PTransport is a map of the virtio-9p device name that corresponds
// to each transport.
var Virtio9PTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-9p-pci",
	TransportCCW:  "virtio-9p-ccw",
	TransportMMIO: "virtio-9p-device",
}

// Valid returns true if the FSDevice structure is valid and complete.
func (fsdev FSDevice) Valid() error {
	if fsdev.ID == "" {
		return fmt.Errorf("FSDevice has empty ID field")
	}
	if fsdev.Path == "" {
		return fmt.Errorf("FSDevice has empty Path field")
	}
	if fsdev.MountTag == "" {
		return fmt.Errorf("FSDevice has empty MountTag field")
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this filesystem device.
func (fsdev FSDevice) QemuParams(config *Config) []string {
	var fsParams []string
	var deviceParams []string
	var qemuParams []string

	deviceParams = append(deviceParams, fsdev.deviceName(config))
	if s := fsdev.Transport.disableModern(config, fsdev.DisableModern); s != "" {
		deviceParams = append(deviceParams, s)
	}
	deviceParams = append(deviceParams, fmt.Sprintf("fsdev=%s", fsdev.ID))
	deviceParams = append(deviceParams, fmt.Sprintf("mount_tag=%s", fsdev.MountTag))
	if fsdev.Transport.isVirtioPCI(config) && fsdev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", fsdev.ROMFile))
	}
	if fsdev.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, ",iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", fsdev.DevNo))
	}

	fsParams = append(fsParams, string(fsdev.FSDriver))
	fsParams = append(fsParams, fmt.Sprintf("id=%s", fsdev.ID))
	fsParams = append(fsParams, fmt.Sprintf("path=%s", fsdev.Path))
	fsParams = append(fsParams, fmt.Sprintf("security_model=%s", fsdev.SecurityModel))

	if fsdev.Multidev != "" {
		fsParams = append(fsParams, fmt.Sprintf("multidevs=%s", fsdev.Multidev))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	qemuParams = append(qemuParams, "-fsdev")
	qemuParams = append(qemuParams, strings.Join(fsParams, ","))

	return qemuParams
}

// deviceName returns the QEMU shared filesystem device name for the current
// combination of driver and transport.
func (fsdev FSDevice) deviceName(config *Config) string {
	if fsdev.Transport == "" {
		fsdev.Transport = fsdev.Transport.defaultTransport(config)
	}

	switch fsdev.Driver {
	case Virtio9P:
		return Virtio9PTransport[fsdev.Transport]
	}

	return string(fsdev.Driver)
}
