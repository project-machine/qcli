package qcli

import "testing"

var (
	deviceBlockString         = "-drive file=/var/lib/vm.img,id=hd0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap,readonly=on -device virtio-blk-pci,drive=hd0,serial=abc-123,disable-modern=true,addr=0x03,bus=pcie.0,logical_block_size=4096,physical_block_size=4096,scsi=off,config-wce=off,romfile=efi-virtio.rom,share-rw=on"
	deviceBlockAddrString     = "-drive file=/var/lib/vm.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x07,bus=pcie.0,scsi=off,config-wce=off"
	deviceBlockPFlashROString = "-drive file=/usr/share/OVMF/OVMF_CODE.fd,id=pflash0,if=pflash,format=raw,readonly=on"
	deviceBlockPFlashRWString = "-drive file=uefi_nvram.fd,id=pflash1,if=pflash,format=raw"
	deviceBlockVirtioCDRom    = "-drive file=ubuntu.iso,id=cdrom0,if=none,format=raw,aio=threads,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off"
	deviceBlockIDECDRom       = "-drive file=ubuntu.iso,id=cdrom0,if=none,format=raw,aio=threads,media=cdrom,readonly=on -device ide-cd,drive=cdrom0,serial=ubuntu.iso,bootindex=0,bus=ide.0"
	deviceBlockSCSIHDStr      = "-drive file=root-disk.qcow,id=drive0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device scsi-hd,drive=drive0,serial=root-disk,bootindex=1,bus=scsi0.0,logical_block_size=512,physical_block_size=512"
	deviceBlockUSBHDStr       = "-drive file=disk0-usb.img,id=drive1,if=none,format=raw,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device usb-storage,drive=drive1,serial=disk0-usb,logical_block_size=512,physical_block_size=512"
)

func TestAppendDeviceBlock(t *testing.T) {
	blkdev := BlockDevice{
		Driver:        VirtioBlock,
		ID:            "hd0",
		File:          "/var/lib/vm.img",
		AIO:           Threads,
		Format:        QCOW2,
		Interface:     NoInterface,
		SCSI:          false,
		WCE:           false,
		DisableModern: true,
		ROMFile:       romfile,
		ShareRW:       true,
		ReadOnly:      true,
		Serial:        "abc-123",
		BlockSize:     4096,
		Cache:         CacheModeUnsafe,
		Discard:       DiscardUnmap,
		DetectZeroes:  DetectZeroesUnmap,
		BusAddr:       "3",
	}
	if blkdev.Transport.isVirtioCCW(nil) {
		blkdev.DevNo = DevNo
	}
	testAppend(blkdev, deviceBlockString, t)
}

func TestAppendDeviceBlockAddr(t *testing.T) {
	blkdev := BlockDevice{
		Driver:    VirtioBlock,
		ID:        "hd0",
		File:      "/var/lib/vm.img",
		Format:    QCOW2,
		Interface: NoInterface,
		BusAddr:   "7",
	}
	if blkdev.Transport.isVirtioCCW(nil) {
		blkdev.DevNo = DevNo
	}
	testAppend(blkdev, deviceBlockAddrString, t)
}

func TestAppendDeviceBlockVirtioCDROM(t *testing.T) {
	blkdev := BlockDevice{
		Driver:    VirtioBlock,
		Interface: NoInterface,
		ID:        "cdrom0",
		AIO:       Threads,
		Serial:    "cdrom0",
		File:      "ubuntu.iso",
		Format:    RAW,
		ReadOnly:  true,
		Media:     "cdrom",
		BootIndex: "0",
	}
	if blkdev.Transport.isVirtioCCW(nil) {
		blkdev.DevNo = DevNo
	}
	testAppend(blkdev, deviceBlockVirtioCDRom, t)
}

func TestAppendDeviceBlockIDECDROM(t *testing.T) {
	blkdev := BlockDevice{
		Driver:    IDECDROM,
		Interface: NoInterface,
		ID:        "cdrom0",
		AIO:       Threads,
		Serial:    "ubuntu.iso",
		File:      "ubuntu.iso",
		Format:    RAW,
		ReadOnly:  true,
		Media:     "cdrom",
		BootIndex: "0",
		Bus:       "ide.0",
	}
	if blkdev.Transport.isVirtioCCW(nil) {
		blkdev.DevNo = DevNo
	}
	testAppend(blkdev, deviceBlockIDECDRom, t)
}

func TestAppendDeviceBlockSCSIHD(t *testing.T) {
	blkdev := BlockDevice{
		Driver:       SCSIHD,
		SCSI:         true,
		Interface:    NoInterface,
		ID:           "drive0",
		AIO:          Threads,
		Serial:       "root-disk",
		File:         "root-disk.qcow",
		Format:       QCOW2,
		BootIndex:    "1",
		Bus:          "scsi0.0",
		Cache:        CacheModeUnsafe,
		Discard:      DiscardUnmap,
		DetectZeroes: DetectZeroesUnmap,
		BlockSize:    512,
	}
	testAppend(blkdev, deviceBlockSCSIHDStr, t)
}

// FIXME: add Scsi + Rotation_rate good/bad tests
// FIXME: add Rotational + Virtio bad test

func TestAppendDeviceBlockPFlashRO(t *testing.T) {
	blkdev := BlockDevice{
		Driver:    PFlash,
		ID:        "pflash0",
		File:      "/usr/share/OVMF/OVMF_CODE.fd",
		Format:    RAW,
		Interface: PFlashInterface,
		ReadOnly:  true,
		DriveOnly: true,
	}
	testAppend(blkdev, deviceBlockPFlashROString, t)
}

func TestAppendDeviceBlockPFlashRW(t *testing.T) {
	blkdev := BlockDevice{
		Driver:    PFlash,
		ID:        "pflash1",
		File:      "uefi_nvram.fd",
		Format:    RAW,
		Interface: PFlashInterface,
		DriveOnly: true,
	}
	testAppend(blkdev, deviceBlockPFlashRWString, t)
}

func TestAppendDeviceBlockUSBHD(t *testing.T) {
	blkdev := BlockDevice{
		Driver:       USBStorage,
		SCSI:         true,
		Interface:    NoInterface,
		ID:           "drive1",
		AIO:          Threads,
		Serial:       "disk0-usb",
		File:         "disk0-usb.img",
		Format:       RAW,
		Cache:        CacheModeUnsafe,
		Discard:      DiscardUnmap,
		DetectZeroes: DetectZeroesUnmap,
		BlockSize:    512,
	}
	testAppend(blkdev, deviceBlockUSBHDStr, t)
}
