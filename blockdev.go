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

type CacheMode string

const (
	CacheModeWriteThrough CacheMode = "writethrough"
	CacheModeWriteBack    CacheMode = "writeback"
	CacheModeNone         CacheMode = "none"
	CacheModeDirectSync   CacheMode = "directsync"
	CacheModeUnsafe       CacheMode = "unsafe"
)

type DetectZeroesMode string

const (
	DetectZeroesOn    DetectZeroesMode = "on"
	DetectZeroesOff   DetectZeroesMode = "off"
	DetectZeroesUnmap DetectZeroesMode = "unmap"
)

type DiscardMode string

const (
	DiscardIgnore DiscardMode = "ignore"
	DiscardUnmap  DiscardMode = "unmap"
)

type FATMode int

const (
	FATMode12 FATMode = 12
	FATMode16 FATMode = 16
	FATMode32 FATMode = 32
)

var FATModes = map[FATMode]bool{
	FATMode12: true,
	FATMode16: true,
	FATMode32: true,
}

// BlockDeviceInterface defines the type of interface the device is connected to.
type BlockDeviceInterface string

// BlockDeviceAIO defines the type of asynchronous I/O the block device should use.
type BlockDeviceAIO string

// BlockDeviceFormat defines the image format used on a block device.
type BlockDeviceFormat string

const (
	// NoInterface for block devices with no interfaces.
	NoInterface BlockDeviceInterface = "none"

	// SCSI represents a SCSI block device interface.
	SCSI BlockDeviceInterface = "scsi"

	PFlashInterface BlockDeviceInterface = "pflash"
)

const (
	// Threads is the pthread asynchronous I/O implementation.
	Threads BlockDeviceAIO = "threads"

	// Native is the pthread asynchronous I/O implementation.
	Native BlockDeviceAIO = "native"
)

const (
	// QCOW2 is the Qemu Copy On Write v2 image format.
	QCOW2 BlockDeviceFormat = "qcow2"
	// RAW is the direct indexing image format
	RAW BlockDeviceFormat = "raw"
)

// BlockDevice represents a qemu block device.
type BlockDevice struct {
	Driver    DeviceDriver         `yaml:"driver"`
	ID        string               `yaml:"id"`
	File      string               `yaml:"file"`
	Interface BlockDeviceInterface `yaml:"interface"`
	AIO       BlockDeviceAIO       `yaml:"aio"`
	Format    BlockDeviceFormat    `yaml:"format"`
	SCSI      bool                 `yaml:"scsi"`
	WCE       bool                 `yaml:"write-cache"`
	BootIndex string               `yaml:"bootindex"`

	// Media is a hint about the what type of content on the disk, e.g media=cdrom
	Media string `yaml:"media"`

	// BlockSize is the linux kernel block {physical,logical}_block_size value
	BlockSize int `yaml:"blocksize-bytes"`

	// RotationRate is the linux kernel block rotation_rate value
	RotationRate int `yaml:"rotation-rate"`

	// BusAddr is the bus address for some block devices (virtio-blk-pci)
	BusAddr string `yaml:"busaddr"`

	Bus string `yaml:"bus"`

	// Serial is the 21-character disk serial value
	Serial string `yaml:"serial"`

	// Cache mode for the disk
	Cache CacheMode `yaml:"cache-mode"`

	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool `yaml:"disable-modern"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"ccw-dev-no"`

	// ShareRW enables multiple qemu instances to share the File
	ShareRW bool `yaml:"share-rw"`

	// ReadOnly sets the block device in readonly mode
	ReadOnly bool `yaml:"read-only"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`

	Discard DiscardMode `yaml:"discard-mode"`

	DetectZeroes DetectZeroesMode `yaml:"detect-zeros-mode"`

	// DriveOnly is a boolean to skip any -device paramters
	// This is currently used for OVMF/UEFI pflash disk only devices
	DriveOnly bool `yaml:"emit-drive-only"`

	// VVFAT driver options
	VVFATDev VVFATDev `yaml:"vvfat-device"`
}

type VVFATDev struct {
	Directory string          `yaml:"dir"`
	Driver    DeviceDriver    `yaml:"driver"`
	FATMode   FATMode         `yaml:"fat-type"` // 12, 16, or 32
	Floppy    bool            `yaml:"floppy"`
	Label     string          `yaml:"label"`
	Transport VirtioTransport `yaml:"transport"`
	ReadWrite bool            `yaml:"rw"` // default read-only
}

func (v VVFATDev) deviceName(config *Config) string {
	if v.Transport == "" {
		v.Transport = v.Transport.defaultTransport(config)
	}

	switch v.Driver {
	case VirtioBlock:
		return VirtioBlockTransport[v.Transport]
	}

	return string(v.Driver)
}

// VirtioBlockTransport is a map of the virtio-blk device name that corresponds
// to each transport.
var VirtioBlockTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-blk-pci",
	TransportCCW:  "virtio-blk-ccw",
	TransportMMIO: "virtio-blk-device",
}

// Valid returns true if the BlockDevice structure is valid and complete.
func (blkdev BlockDevice) Valid() error {

	if blkdev.ID == "" {
		return fmt.Errorf("BlockDevice missing ID")
	}
	if blkdev.Driver == "" {
		return fmt.Errorf("BlockDevice ID=%s missing Driver", blkdev.ID)
	}
	switch blkdev.Driver {
	case VVFAT:
		if blkdev.VVFATDev.Directory == "" {
			return fmt.Errorf("BlockDevice ID=%s VVFAT missing required Directory", blkdev.ID)
		}
		if ok := FATModes[blkdev.VVFATDev.FATMode]; !ok {
			return fmt.Errorf("BlockDevice ID=%s VVFAT invalid FATMode %d", blkdev.ID, blkdev.VVFATDev.FATMode)
		}
	default:
		if blkdev.File == "" {
			return fmt.Errorf("BlockDevice ID=%s missing File", blkdev.ID)
		}
		if blkdev.Interface == "" {
			return fmt.Errorf("BlockDevice ID=%s missing Interface", blkdev.ID)
		}
		if blkdev.Format == "" {
			return fmt.Errorf("BlockDevice ID=%s missing Format", blkdev.ID)
		}
		if blkdev.RotationRate > 0 && strings.HasPrefix(string(blkdev.Driver), "virtio") {
			return fmt.Errorf("BlockDevice ID=%s with RotationRate cannot be Driver=virtio*", blkdev.ID)
		}
	}
	return nil
}

// FIXME: this should use -blockdev, instead of -drive
// QemuParams returns the qemu parameters built out of this block device.
func (blkdev BlockDevice) QemuParams(config *Config) []string {
	var driveParams []string
	var blockdevParams []string
	var deviceParams []string
	var qemuParams []string

	switch blkdev.Driver {
	case VVFAT:
		blockdevParams = append(blockdevParams, fmt.Sprintf("driver=%s", blkdev.deviceName(config)))
		blockdevParams = append(blockdevParams, fmt.Sprintf("node-name=%s", blkdev.ID))
		blockdevParams = append(blockdevParams, fmt.Sprintf("dir=%s", blkdev.VVFATDev.Directory))
		if blkdev.VVFATDev.FATMode > 0 {
			blockdevParams = append(blockdevParams, fmt.Sprintf("fat-type=%d", blkdev.VVFATDev.FATMode))
		} else {
			blockdevParams = append(blockdevParams, "fat-type=32")
		}

		if blkdev.VVFATDev.Floppy {
			blockdevParams = append(blockdevParams, "floppy=on")
		} else {
			blockdevParams = append(blockdevParams, "floppy=off")
		}

		if blkdev.VVFATDev.Label != "" {
			blockdevParams = append(blockdevParams, fmt.Sprintf("label=%s", blkdev.VVFATDev.Label))
		}

		if blkdev.VVFATDev.ReadWrite {
			blockdevParams = append(blockdevParams, "read-only=off")
		} else {
			blockdevParams = append(blockdevParams, "read-only=on")
		}

		deviceParams = append(deviceParams, blkdev.VVFATDev.deviceName(config))
		deviceParams = append(deviceParams, fmt.Sprintf("drive=%s", blkdev.ID))

		qemuParams = append(qemuParams, "-blockdev")
		qemuParams = append(qemuParams, strings.Join(blockdevParams, ","))

	default:
		// drive parameters
		driveParams = append(driveParams, fmt.Sprintf("file=%s", blkdev.File))
		driveParams = append(driveParams, fmt.Sprintf("id=%s", blkdev.ID))
		driveParams = append(driveParams, fmt.Sprintf("if=%s", blkdev.Interface))
		driveParams = append(driveParams, fmt.Sprintf("format=%s", blkdev.Format))

		if blkdev.AIO != "" {
			driveParams = append(driveParams, fmt.Sprintf("aio=%s", blkdev.AIO))
		}

		if blkdev.Cache != "" {
			driveParams = append(driveParams, fmt.Sprintf("cache=%s", blkdev.Cache))
		}

		if blkdev.Discard != "" {
			driveParams = append(driveParams, fmt.Sprintf("discard=%s", blkdev.Discard))
		}

		if blkdev.DetectZeroes != "" {
			driveParams = append(driveParams, fmt.Sprintf("detect-zeroes=%s", blkdev.DetectZeroes))
		}

		if blkdev.Media != "" {
			driveParams = append(driveParams, fmt.Sprintf("media=%s", blkdev.Media))
		}

		if blkdev.ReadOnly {
			driveParams = append(driveParams, "readonly=on")
		}

		qemuParams = append(qemuParams, "-drive")
		qemuParams = append(qemuParams, strings.Join(driveParams, ","))

		// for DriveOnly blockdev devices, no need for -device params
		if blkdev.DriveOnly {
			return qemuParams
		}

		// All device parameters must be after DriveOnly
		deviceParams = append(deviceParams, blkdev.deviceName(config))
		deviceParams = append(deviceParams, fmt.Sprintf("drive=%s", blkdev.ID))
		if blkdev.Serial != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("serial=%s", blkdev.Serial))
		} else {
			deviceParams = append(deviceParams, fmt.Sprintf("serial=%s", blkdev.ID))
		}

		if blkdev.BootIndex != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("bootindex=%s", blkdev.BootIndex))
		}

		if blkdev.Driver == VirtioBlock {
			if s := blkdev.Transport.disableModern(config, blkdev.DisableModern); s != "" {
				deviceParams = append(deviceParams, s)
			}

			// virtio can have a BusAddr since they are pci devices
			addr := config.pciBusSlots.GetSlot(blkdev.BusAddr)
			if addr > 0 {
				deviceParams = append(deviceParams, fmt.Sprintf("addr=0x%02x", addr))
				bus := "pcie.0"
				if blkdev.Bus != "" {
					bus = blkdev.Bus
				}
				deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", bus))
			}
		}

		if blkdev.Driver == SCSIHD && blkdev.Bus != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", blkdev.Bus))
		}

		if blkdev.Driver == IDECDROM {
			bus := "ide.0"
			if blkdev.Bus != "" {
				bus = blkdev.Bus
			}
			deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", bus))
		}

		if blkdev.RotationRate > 0 && !strings.HasPrefix(string(blkdev.Driver), "virtio") {
			deviceParams = append(deviceParams, fmt.Sprintf("rotation_rate=%d", blkdev.RotationRate))
		}

		if blkdev.BlockSize > 0 {
			deviceParams = append(deviceParams, fmt.Sprintf("logical_block_size=%d", blkdev.BlockSize))
			deviceParams = append(deviceParams, fmt.Sprintf("physical_block_size=%d", blkdev.BlockSize))
		}

		if !blkdev.SCSI && blkdev.Driver != IDECDROM {
			deviceParams = append(deviceParams, "scsi=off")
		}

		if !blkdev.WCE && blkdev.Driver == VirtioBlock {
			deviceParams = append(deviceParams, "config-wce=off")
		}

		if blkdev.Transport.isVirtioPCI(config) && blkdev.ROMFile != "" {
			deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", blkdev.ROMFile))
		}

		if blkdev.Transport.isVirtioCCW(config) {
			deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", blkdev.DevNo))
		}

		if blkdev.ShareRW {
			deviceParams = append(deviceParams, "share-rw=on")
		}
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (blkdev BlockDevice) deviceName(config *Config) string {
	if blkdev.Transport == "" {
		blkdev.Transport = blkdev.Transport.defaultTransport(config)
	}

	switch blkdev.Driver {
	case VirtioBlock:
		return VirtioBlockTransport[blkdev.Transport]
	}

	return string(blkdev.Driver)
}
