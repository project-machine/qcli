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

// ObjectType is a string representing a qemu object type.
type ObjectType string

const (
	// MemoryBackendFile represents a guest memory mapped file.
	MemoryBackendFile ObjectType = "memory-backend-file"

	// MemoryBackendEPC represents a guest memory backend EPC for SGX.
	MemoryBackendEPC ObjectType = "memory-backend-epc"

	// TDXGuest represents a TDX object
	TDXGuest ObjectType = "tdx-guest"

	// SEVGuest represents an SEV guest object
	SEVGuest ObjectType = "sev-guest"

	// SecExecGuest represents an s390x Secure Execution (Protected Virtualization in QEMU) object
	SecExecGuest ObjectType = "s390-pv-guest"
	// PEFGuest represent ppc64le PEF(Protected Execution Facility) object.
	PEFGuest ObjectType = "pef-guest"

	LegacyMemPath ObjectType = "legacy-mem-path"
)

// Object is a qemu object representation.
type Object struct {
	// Driver is the qemu device driver
	Driver DeviceDriver `yaml:"driver"`

	// Type is the qemu object type.
	Type ObjectType `yaml:"type"`

	// ID is the user defined object ID.
	ID string `yaml:"id"`

	// DeviceID is the user defined device ID.
	DeviceID string `yaml:"device-id"`

	// MemPath is the object's memory path.
	// This is only relevant for memory objects
	MemPath string `yaml:"mem-path"`

	// Size is the object size in bytes
	Size uint64 `yaml:"size-bytes"`

	// Debug this is a debug object
	Debug bool `yaml:"debug-enable"`

	// File is the device file
	File string `yaml:"file"`

	// CBitPos is the location of the C-bit in a guest page table entry
	// This is only relevant for sev-guest objects
	CBitPos uint32 `yaml:"c-bit-position"`

	// ReducedPhysBits is the reduction in the guest physical address space
	// This is only relevant for sev-guest objects
	ReducedPhysBits uint32 `yaml:"reduce-phys-bits"`

	// ReadOnly specifies whether `MemPath` is opened read-only or read/write (default)
	ReadOnly bool `yaml:"read-only"`

	// Prealloc enables memory preallocation
	Prealloc bool `yaml:"pre-allocate"`
}

// Valid returns true if the Object structure is valid and complete.
func (object Object) Valid() bool {
	switch object.Type {
	case MemoryBackendFile:
		return object.ID != "" && object.MemPath != "" && object.Size != 0
	case MemoryBackendEPC:
		return object.ID != "" && object.Size != 0
	case TDXGuest:
		return object.ID != "" && object.File != "" && object.DeviceID != ""
	case SEVGuest:
		return object.ID != "" && object.File != "" && object.CBitPos != 0 && object.ReducedPhysBits != 0
	case SecExecGuest:
		return object.ID != ""
	case PEFGuest:
		return object.ID != "" && object.File != ""
	case LegacyMemPath:
		panic("LegacyMemPath")
		return object.MemPath != ""

	default:
		return false
	}
}

// QemuParams returns the qemu parameters built out of this Object device.
func (object Object) QemuParams(config *Config) []string {
	var objectParams []string
	var deviceParams []string
	var driveParams []string
	var qemuParams []string

	switch object.Type {
	case LegacyMemPath:
		qemuParams = append(qemuParams, "-mem-path", object.MemPath)
		if object.Prealloc {
			qemuParams = append(qemuParams, "-mem-prealloc")
		}
	case MemoryBackendFile:
		objectParams = append(objectParams, string(object.Type))
		objectParams = append(objectParams, fmt.Sprintf("id=%s", object.ID))
		objectParams = append(objectParams, fmt.Sprintf("mem-path=%s", object.MemPath))
		objectParams = append(objectParams, fmt.Sprintf("size=%d", object.Size))

		deviceParams = append(deviceParams, string(object.Driver))
		deviceParams = append(deviceParams, fmt.Sprintf("id=%s", object.DeviceID))
		deviceParams = append(deviceParams, fmt.Sprintf("memdev=%s", object.ID))

		if object.ReadOnly {
			objectParams = append(objectParams, "readonly=on")
			deviceParams = append(deviceParams, "unarmed=on")
		}
	case MemoryBackendEPC:
		objectParams = append(objectParams, string(object.Type))
		objectParams = append(objectParams, fmt.Sprintf("id=%s", object.ID))
		objectParams = append(objectParams, fmt.Sprintf("size=%d", object.Size))
		if object.Prealloc {
			objectParams = append(objectParams, "prealloc=on")
		}

	case TDXGuest:
		objectParams = append(objectParams, string(object.Type))
		objectParams = append(objectParams, fmt.Sprintf("id=%s", object.ID))
		if object.Debug {
			objectParams = append(objectParams, "debug=on")
		}
		deviceParams = append(deviceParams, string(object.Driver))
		deviceParams = append(deviceParams, fmt.Sprintf("id=%s", object.DeviceID))
		deviceParams = append(deviceParams, fmt.Sprintf("file=%s", object.File))
	case SEVGuest:
		objectParams = append(objectParams, string(object.Type))
		objectParams = append(objectParams, fmt.Sprintf("id=%s", object.ID))
		objectParams = append(objectParams, fmt.Sprintf("cbitpos=%d", object.CBitPos))
		objectParams = append(objectParams, fmt.Sprintf("reduced-phys-bits=%d", object.ReducedPhysBits))

		driveParams = append(driveParams, "if=pflash,format=raw,readonly=on")
		driveParams = append(driveParams, fmt.Sprintf("file=%s", object.File))
	case SecExecGuest:
		objectParams = append(objectParams, string(object.Type))
		objectParams = append(objectParams, fmt.Sprintf("id=%s", object.ID))
	case PEFGuest:
		objectParams = append(objectParams, string(object.Type))
		objectParams = append(objectParams, fmt.Sprintf("id=%s", object.ID))

		deviceParams = append(deviceParams, string(object.Driver))
		deviceParams = append(deviceParams, fmt.Sprintf("id=%s", object.DeviceID))
		deviceParams = append(deviceParams, fmt.Sprintf("host-path=%s", object.File))

	}

	if len(deviceParams) > 0 {
		qemuParams = append(qemuParams, "-device")
		qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	}

	if len(objectParams) > 0 {
		qemuParams = append(qemuParams, "-object")
		qemuParams = append(qemuParams, strings.Join(objectParams, ","))
	}

	if len(driveParams) > 0 {
		qemuParams = append(qemuParams, "-drive")
		qemuParams = append(qemuParams, strings.Join(driveParams, ","))
	}

	return qemuParams
}
