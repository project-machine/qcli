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

// SCSIController represents a SCSI controller device.
type SCSIControllerDevice struct {
	ID string `yaml:"id"`

	// Bus on which the SCSI controller is attached, this is optional
	Bus string `yaml:"bus,omitempty"`

	// Addr is the PCI address offset, this is optional
	Addr string `yaml:"addr,omitempty"`

	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool `yaml:"disable-modern,omitempty"`

	// IOThread is the IO thread on which IO will be handled
	IOThread string `yaml:"iothread,omitempty"`

	// IOThread object tunables
	IOThreadPoll   int `yaml:"iothread-poll,omitempty"`
	IOThreadMaxNS  int `yaml:"iothread-max-ns,omitempty"`
	IOThreadShrink int `yaml:"iothread-shrink,omitempty"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"romfile,omitempty"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"devno,omitempty"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport
}

// SCSIControllerTransport is a map of the virtio-scsi device name that
// corresponds to each transport.
var SCSIControllerTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-scsi-pci",
	TransportCCW:  "virtio-scsi-ccw",
	TransportMMIO: "virtio-scsi-device",
}

// Valid returns true if the SCSIController structure is valid and complete.
func (scsiCon SCSIControllerDevice) Valid() error {
	if scsiCon.ID == "" {
		return fmt.Errorf("SCSIController has empty ID field")
	}
	return nil
}

// QemuParams returns the qemu parameters built out of this SCSIController device.
func (scsiCon SCSIControllerDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string
	var objectParams []string

	driver := scsiCon.deviceName(config)
	deviceParams = append(deviceParams, fmt.Sprintf("%s,id=%s", driver, scsiCon.ID))
	addr := config.pciBusSlots.GetSlot(scsiCon.Addr)
	if addr > 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("addr=0x%02x", addr))
		bus := "pcie.0"
		if scsiCon.Bus != "" {
			bus = scsiCon.Bus
		}
		deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", bus))
	}
	if s := scsiCon.Transport.disableModern(config, scsiCon.DisableModern); s != "" {
		deviceParams = append(deviceParams, s)
	}
	if scsiCon.IOThread != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("iothread=%s", scsiCon.IOThread))
		// FIXME, add in tuneables
		objectParams = append(objectParams, fmt.Sprintf("iothread,poll-max-ns=32,id=%s", scsiCon.IOThread))
	}
	if scsiCon.Transport.isVirtioPCI(config) && scsiCon.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", scsiCon.ROMFile))
	}

	if scsiCon.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, "iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", scsiCon.DevNo))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	if len(objectParams) > 0 {
		qemuParams = append(qemuParams, "-object")
		qemuParams = append(qemuParams, strings.Join(objectParams, ","))
	}
	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (scsiCon SCSIControllerDevice) deviceName(config *Config) string {
	if scsiCon.Transport == "" {
		scsiCon.Transport = scsiCon.Transport.defaultTransport(config)
	}

	return SCSIControllerTransport[scsiCon.Transport]
}
