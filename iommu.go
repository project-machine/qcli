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

import "strings"

// IommuDev represents a Intel IOMMU Device
type IommuDev struct {
	Intremap    bool `yaml:"interupt-remap"`
	DeviceIotlb bool `yaml:"device-iotlb"`
	CachingMode bool `yaml:"caching-mode"`
}

// Valid returns true if the IommuDev is valid
func (dev IommuDev) Valid() error {
	return nil
}

// deviceName the qemu device name
func (dev IommuDev) deviceName() string {
	return "intel-iommu"
}

// QemuParams returns the qemu parameters built out of the IommuDev.
func (dev IommuDev) QemuParams(_ *Config) []string {
	var qemuParams []string
	var deviceParams []string

	deviceParams = append(deviceParams, dev.deviceName())
	if dev.Intremap {
		deviceParams = append(deviceParams, "intremap=on")
	} else {
		deviceParams = append(deviceParams, "intremap=off")
	}

	if dev.DeviceIotlb {
		deviceParams = append(deviceParams, "device-iotlb=on")
	} else {
		deviceParams = append(deviceParams, "device-iotlb=off")
	}

	if dev.CachingMode {
		deviceParams = append(deviceParams, "caching-mode=on")
	} else {
		deviceParams = append(deviceParams, "caching-mode=off")
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	return qemuParams
}
