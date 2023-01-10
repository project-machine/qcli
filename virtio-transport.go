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

import "runtime"

// VirtioTransport is the transport in use for a virtio device.
type VirtioTransport string

const (
	// TransportPCI is the PCI transport for virtio device.
	TransportPCI VirtioTransport = "pci"

	// TransportCCW is the CCW transport for virtio devices.
	TransportCCW VirtioTransport = "ccw"

	// TransportMMIO is the MMIO transport for virtio devices.
	TransportMMIO VirtioTransport = "mmio"
)

// defaultTransport returns the default transport for the current combination
// of host's architecture and QEMU machine type.
func (transport VirtioTransport) defaultTransport(config *Config) VirtioTransport {
	switch runtime.GOARCH {
	case "amd64", "386":
		if config != nil && config.Machine.Type == MachineTypeMicrovm {
			return TransportMMIO
		}
		return TransportPCI
	case "s390x":
		return TransportCCW
	default:
		return TransportPCI
	}
}

// isVirtioPCI returns true if the transport is PCI.
func (transport VirtioTransport) isVirtioPCI(config *Config) bool {
	if transport == "" {
		transport = transport.defaultTransport(config)
	}

	return transport == TransportPCI
}

// isVirtioCCW returns true if the transport is CCW.
func (transport VirtioTransport) isVirtioCCW(config *Config) bool {
	if transport == "" {
		transport = transport.defaultTransport(config)
	}

	return transport == TransportCCW
}

// getName returns the name of the current transport.
func (transport VirtioTransport) getName(config *Config) string {
	if transport == "" {
		transport = transport.defaultTransport(config)
	}

	return string(transport)
}

// disableModern returns the parameters with the disable-modern option.
// In case the device driver is not a PCI device and it doesn't have the option
// an empty string is returned.
func (transport VirtioTransport) disableModern(config *Config, disable bool) string {
	if !transport.isVirtioPCI(config) {
		return ""
	}

	if disable {
		return "disable-modern=true"
	}

	return "disable-modern=false"
}
