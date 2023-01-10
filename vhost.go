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

// VhostUserDevice represents a qemu vhost-user device meant to be passed
// in to the guest
type VhostUserDevice struct {
	SocketPath     string //path to vhostuser socket on host
	CharDevID      string
	TypeDevID      string //variable QEMU parameter based on value of VhostUserType
	Address        string //used for MAC address in net case
	Tag            string //virtio-fs volume id for mounting inside guest
	CacheSize      uint32 //virtio-fs DAX cache size in MiB
	SharedVersions bool   //enable virtio-fs shared version metadata
	VhostUserType  DeviceDriver

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string

	// DevNo identifies the CCW device for s390x.
	DevNo string

	// Transport is the virtio transport for this device.
	Transport VirtioTransport
}

// VhostUserNetTransport is a map of the virtio-net device name that
// corresponds to each transport.
var VhostUserNetTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-net-pci",
	TransportCCW:  "virtio-net-ccw",
	TransportMMIO: "virtio-net-device",
}

// VhostUserSCSITransport is a map of the vhost-user-scsi device name that
// corresponds to each transport.
var VhostUserSCSITransport = map[VirtioTransport]string{
	TransportPCI:  "vhost-user-scsi-pci",
	TransportCCW:  "vhost-user-scsi-ccw",
	TransportMMIO: "vhost-user-scsi-device",
}

// VhostUserBlkTransport is a map of the vhost-user-blk device name that
// corresponds to each transport.
var VhostUserBlkTransport = map[VirtioTransport]string{
	TransportPCI:  "vhost-user-blk-pci",
	TransportCCW:  "vhost-user-blk-ccw",
	TransportMMIO: "vhost-user-blk-device",
}

// VhostUserFSTransport is a map of the vhost-user-fs device name that
// corresponds to each transport.
var VhostUserFSTransport = map[VirtioTransport]string{
	TransportPCI:  "vhost-user-fs-pci",
	TransportCCW:  "vhost-user-fs-ccw",
	TransportMMIO: "vhost-user-fs-device",
}

// Valid returns true if there is a valid structure defined for VhostUserDevice
func (vhostuserDev VhostUserDevice) Valid() error {

	if vhostuserDev.SocketPath == "" {
		return fmt.Errorf("VhostUserDevice has empty SocketPath field")
	}
	if vhostuserDev.CharDevID == "" {
		return fmt.Errorf("VhostUserDevice has empty CharDevID field")
	}

	switch vhostuserDev.VhostUserType {
	case VhostUserNet, VhostUserSCSI, VhostUserBlk, VhostUserFS:
		break
	default:
		return fmt.Errorf("VhostUserDevice has unknown VhostUserType: %s", vhostuserDev.VhostUserType)
	}

	if vhostuserDev.VhostUserType == VhostUserNet {
		if vhostuserDev.TypeDevID == "" {
			return fmt.Errorf("VhostUserDevice Type=VhostUserNet has empty TypeDevID field")
		}
		if vhostuserDev.Address == "" {
			return fmt.Errorf("VhostUserDevice Type=VhostUserNet has empty Address field")
		}
	}
	if vhostuserDev.VhostUserType == VhostUserSCSI {
		if vhostuserDev.TypeDevID == "" {
			return fmt.Errorf("VhostUserDevice Type=VhostUserSCSI has empty TypeDevID field")
		}
	}
	if vhostuserDev.VhostUserType == VhostUserFS {
		if vhostuserDev.Tag == "" {
			return fmt.Errorf("VhostUserDevice Type=VhostUserFS has empty Tag field")
		}
	}

	return nil
}

// QemuNetParams builds QEMU netdev and device parameters for a VhostUserNet device
func (vhostuserDev VhostUserDevice) QemuNetParams(config *Config) []string {
	var qemuParams []string
	var netParams []string
	var deviceParams []string

	driver := vhostuserDev.deviceName(config)
	if driver == "" {
		return nil
	}

	netParams = append(netParams, "type=vhost-user")
	netParams = append(netParams, fmt.Sprintf("id=%s", vhostuserDev.TypeDevID))
	netParams = append(netParams, fmt.Sprintf("chardev=%s", vhostuserDev.CharDevID))
	netParams = append(netParams, "vhostforce")

	deviceParams = append(deviceParams, driver)
	deviceParams = append(deviceParams, fmt.Sprintf("netdev=%s", vhostuserDev.TypeDevID))
	deviceParams = append(deviceParams, fmt.Sprintf("mac=%s", vhostuserDev.Address))

	if vhostuserDev.Transport.isVirtioPCI(config) && vhostuserDev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", vhostuserDev.ROMFile))
	}

	qemuParams = append(qemuParams, "-netdev")
	qemuParams = append(qemuParams, strings.Join(netParams, ","))
	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// QemuSCSIParams builds QEMU device parameters for a VhostUserSCSI device
func (vhostuserDev VhostUserDevice) QemuSCSIParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	driver := vhostuserDev.deviceName(config)
	if driver == "" {
		return nil
	}

	deviceParams = append(deviceParams, driver)
	deviceParams = append(deviceParams, fmt.Sprintf("id=%s", vhostuserDev.TypeDevID))
	deviceParams = append(deviceParams, fmt.Sprintf("chardev=%s", vhostuserDev.CharDevID))

	if vhostuserDev.Transport.isVirtioPCI(config) && vhostuserDev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", vhostuserDev.ROMFile))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// QemuBlkParams builds QEMU device parameters for a VhostUserBlk device
func (vhostuserDev VhostUserDevice) QemuBlkParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	driver := vhostuserDev.deviceName(config)
	if driver == "" {
		return nil
	}

	deviceParams = append(deviceParams, driver)
	deviceParams = append(deviceParams, "logical_block_size=4096")
	deviceParams = append(deviceParams, "size=512M")
	deviceParams = append(deviceParams, fmt.Sprintf("chardev=%s", vhostuserDev.CharDevID))

	if vhostuserDev.Transport.isVirtioPCI(config) && vhostuserDev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", vhostuserDev.ROMFile))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// QemuFSParams builds QEMU device parameters for a VhostUserFS device
func (vhostuserDev VhostUserDevice) QemuFSParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	driver := vhostuserDev.deviceName(config)
	if driver == "" {
		return nil
	}

	deviceParams = append(deviceParams, driver)
	deviceParams = append(deviceParams, fmt.Sprintf("chardev=%s", vhostuserDev.CharDevID))
	deviceParams = append(deviceParams, fmt.Sprintf("tag=%s", vhostuserDev.Tag))
	if vhostuserDev.CacheSize != 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("cache-size=%dM", vhostuserDev.CacheSize))
	}
	if vhostuserDev.SharedVersions {
		deviceParams = append(deviceParams, "versiontable=/dev/shm/fuse_shared_versions")
	}
	if vhostuserDev.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, "iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", vhostuserDev.DevNo))
	}
	if vhostuserDev.Transport.isVirtioPCI(config) && vhostuserDev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", vhostuserDev.ROMFile))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// QemuParams returns the qemu parameters built out of this vhostuser device.
func (vhostuserDev VhostUserDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var charParams []string
	var deviceParams []string

	charParams = append(charParams, "socket")
	charParams = append(charParams, fmt.Sprintf("id=%s", vhostuserDev.CharDevID))
	charParams = append(charParams, fmt.Sprintf("path=%s", vhostuserDev.SocketPath))

	qemuParams = append(qemuParams, "-chardev")
	qemuParams = append(qemuParams, strings.Join(charParams, ","))

	switch vhostuserDev.VhostUserType {
	case VhostUserNet:
		deviceParams = vhostuserDev.QemuNetParams(config)
	case VhostUserSCSI:
		deviceParams = vhostuserDev.QemuSCSIParams(config)
	case VhostUserBlk:
		deviceParams = vhostuserDev.QemuBlkParams(config)
	case VhostUserFS:
		deviceParams = vhostuserDev.QemuFSParams(config)
	default:
		return nil
	}

	if deviceParams != nil {
		return append(qemuParams, deviceParams...)
	}

	return nil
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (vhostuserDev VhostUserDevice) deviceName(config *Config) string {
	if vhostuserDev.Transport == "" {
		vhostuserDev.Transport = vhostuserDev.Transport.defaultTransport(config)
	}

	switch vhostuserDev.VhostUserType {
	case VhostUserNet:
		return VhostUserNetTransport[vhostuserDev.Transport]
	case VhostUserSCSI:
		return VhostUserSCSITransport[vhostuserDev.Transport]
	case VhostUserBlk:
		return VhostUserBlkTransport[vhostuserDev.Transport]
	case VhostUserFS:
		return VhostUserFSTransport[vhostuserDev.Transport]
	default:
		return ""
	}
}
