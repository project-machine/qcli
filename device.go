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
	"reflect"
	"strings"
)

// Device is the qemu device interface.
type Device interface {
	Valid() error
	QemuParams(config *Config) []string
}

// DeviceDriver is the device driver string.
type DeviceDriver string

const (
	// LegacySerial is the legacy serial device driver
	LegacySerial DeviceDriver = "serial"

	// NVDIMM is the Non Volatile DIMM device driver.
	NVDIMM DeviceDriver = "nvdimm"

	// VirtioNet is the virtio networking device driver.
	VirtioNet DeviceDriver = "virtio-net"

	// VirtioNetPCI is the virt-io pci networking device driver.
	VirtioNetPCI DeviceDriver = "virtio-net-pci"

	// VirtioNetCCW is the virt-io ccw networking device driver.
	VirtioNetCCW DeviceDriver = "virtio-net-ccw"

	// E1000 is the emulated Intel E1000 networking device driver
	E1000 DeviceDriver = "e1000"

	// VirtioBlock is the block device driver.
	VirtioBlock DeviceDriver = "virtio-blk"

	// IDEHardDisk is the block device driver
	IDEHardDisk DeviceDriver = "ide-hd"

	// IDECDROM is the block device driver
	IDECDROM DeviceDriver = "ide-cd"

	// SCSIHD is the block device driver
	SCSIHD DeviceDriver = "scsi-hd"

	// SCSICD is the block device driver
	SCSICD DeviceDriver = "scsi-cd"

	// NVME is the block device driver
	NVME DeviceDriver = "nvme"

	// USBStorage is the block device driver
	USBStorage DeviceDriver = "usb-storage"

	// Console is the console device driver.
	Console DeviceDriver = "virtconsole"

	// Virtio9P is the 9pfs device driver.
	Virtio9P DeviceDriver = "virtio-9p"

	// VirtioScsi is the scsi controller over virtio driver
	VirtioScsi DeviceDriver = "virtio-scsi-pci"

	// VirtioSerial is the serial device driver.
	VirtioSerial DeviceDriver = "virtio-serial"

	// VirtioSerialPort is the serial port device driver.
	VirtioSerialPort DeviceDriver = "virtserialport"

	// VirtioRng is the paravirtualized RNG device driver.
	VirtioRng DeviceDriver = "virtio-rng"

	// VirtioRngPCI is the paravirtualized RNG device driver on PCI bus
	VirtioRngPCI DeviceDriver = "virtio-rng-pci"

	// VirtioRngPCI is the paravirtualized RNG device driver on CCW bus
	VirtioRngCCW DeviceDriver = "virtio-rng-ccw"

	// VirtioBalloon is the memory balloon device driver.
	VirtioBalloon DeviceDriver = "virtio-balloon"

	//VhostUserSCSI represents a SCSI vhostuser device type.
	VhostUserSCSI DeviceDriver = "vhost-user-scsi"

	//VhostUserNet represents a net vhostuser device type.
	VhostUserNet DeviceDriver = "virtio-net"

	//VhostUserBlk represents a block vhostuser device type.
	VhostUserBlk DeviceDriver = "vhost-user-blk"

	//VhostUserFS represents a virtio-fs vhostuser device type
	VhostUserFS DeviceDriver = "vhost-user-fs"

	// PCIBridgeDriver represents a PCI bridge device type.
	PCIBridgeDriver DeviceDriver = "pci-bridge"

	// PCIePCIBridgeDriver represents a PCIe to PCI bridge device type.
	PCIePCIBridgeDriver DeviceDriver = "pcie-pci-bridge"

	// VfioPCI is the vfio driver with PCI transport.
	VfioPCI DeviceDriver = "vfio-pci"

	// VfioCCW is the vfio driver with CCW transport.
	VfioCCW DeviceDriver = "vfio-ccw"

	// VfioAP is the vfio driver with AP transport.
	VfioAP DeviceDriver = "vfio-ap"

	// VHostVSockPCI is a generic Vsock vhost device with PCI transport.
	VHostVSockPCI DeviceDriver = "vhost-vsock-pci"

	// PCIeRootPort is a PCIe Root Port, the PCIe device should be hotplugged to this port.
	PCIeRootPort DeviceDriver = "pcie-root-port"

	// Loader is the Loader device driver.
	Loader DeviceDriver = "loader"

	// SpaprTPMProxy is used for enabling guest to run in secure mode on ppc64le.
	SpaprTPMProxy DeviceDriver = "spapr-tpm-proxy"

	// PFlash
	PFlash DeviceDriver = "pflash"

	// USB-XHCI-Controller
	USBXHCIController DeviceDriver = "qemu-xhci"

	// AHCI ICH9 Controller
	ICH9AHCIController DeviceDriver = "ich9-ahci"

	// PIIX3 IDE Controller
	PIIX3IDEController DeviceDriver = "piix3-ide"

	// PIIX4 IDE Controller
	PIIX4IDEController DeviceDriver = "piix4-ide"

	// TPM-TIS TPM Device
	TPMTISDevice DeviceDriver = "tpm-tis"

	// TPM-CRB TPM Device
	TPMCRBDebice DeviceDriver = "tpm-crb"

	// PCI Serial Device
	PCISerialDevice DeviceDriver = "pci-serial"
)

func (config *Config) appendDevices() error {
	// I'd really like to keep the Devices []Device but unmarshaling it is a
	// huge page, so we'll have a list of each device type in the config and
	// sort through each devices list and append if valid.

	// FIXME: if I could invoke the fields on config that match a regex then
	// we could have a single switch case which matches .+Devices and then
	// appends each device to config.devices.
	fields := reflect.VisibleFields(reflect.TypeOf(Config{}))

	// insert pci and scsi controllers first
	for _, field := range fields {
		switch field.Name {
		case "PCIeRootPortDevices":
			for _, d := range config.PCIeRootPortDevices {
				config.devices = append(config.devices, d)
			}
		case "SCSIControllerDevices": // controllers have to be before blkdev
			for _, d := range config.SCSIControllerDevices {
				config.devices = append(config.devices, d)
			}
		case "IDEControllerDevices": // controllers have to be before blkdev
			for _, d := range config.IDEControllerDevices {
				config.devices = append(config.devices, d)
			}
		case "USBControllerDevices": // controllers have to be before blkdev
			for _, d := range config.USBControllerDevices {
				config.devices = append(config.devices, d)
			}
		}
	}

	// insert the remaining devices
	for _, field := range fields {
		switch field.Name {
		case "BlkDevices":
			for _, d := range config.BlkDevices {
				config.devices = append(config.devices, d)
			}
		case "CharDevices":
			for _, d := range config.CharDevices {
				config.devices = append(config.devices, d)
			}
		case "LegacySerialDevices":
			for _, d := range config.LegacySerialDevices {
				config.devices = append(config.devices, d)
			}
		case "MonitorDevices":
			for _, d := range config.MonitorDevices {
				config.devices = append(config.devices, d)
			}
		case "NetDevices":
			for _, d := range config.NetDevices {
				config.devices = append(config.devices, d)
			}
		case "RngDevices":
			for _, d := range config.RngDevices {
				config.devices = append(config.devices, d)
			}
		case "SerialDevices":
			for _, d := range config.SerialDevices {
				config.devices = append(config.devices, d)
			}
		case "UEFIFirmwareDevices":
			for _, d := range config.UEFIFirmwareDevices {
				config.devices = append(config.devices, d)
			}
		}
	}

	var errors []string
	for _, d := range config.devices {
		if err := d.Valid(); err != nil {
			errors = append(errors, err.Error())
			continue
		}

		config.qemuParams = append(config.qemuParams, d.QemuParams(config)...)
	}

	if len(errors) > 0 {
		return fmt.Errorf("Failed to append %d devices: %s", len(errors), strings.Join(errors, ", "))
	}

	return nil
}
