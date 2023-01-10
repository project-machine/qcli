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

// PVPanicDevice represents a qemu pvpanic device.
type PVPanicDevice struct {
	NoShutdown bool `yaml:"no-shutdown-enable"`
}

// Valid always returns true for pvpanic device
func (dev PVPanicDevice) Valid() error {
	return nil
}

// QemuParams returns the qemu parameters built out of this serial device.
func (dev PVPanicDevice) QemuParams(config *Config) []string {
	if dev.NoShutdown {
		return []string{"-device", "pvpanic", "-no-shutdown"}
	}
	return []string{"-device", "pvpanic"}
}
