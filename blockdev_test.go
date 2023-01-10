package qcli

import "testing"

var (
	deviceBlockString         = "-drive file=/var/lib/vm.img,id=hd0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap,readonly=on -device virtio-blk-pci,drive=hd0,serial=abc-123,bootindex=0,disable-modern=true,logical_block_size=4096,physical_block_size=4096,scsi=off,config-wce=off,romfile=efi-virtio.rom,share-rw=on"
	deviceBlockPFlashROString = "-drive file=/usr/share/OVMF/OVMF_CODE.fd,id=pflash0,if=pflash,format=raw,readonly=on"
	deviceBlockPFlashRWString = "-drive file=uefi_nvram.fd,id=pflash1,if=pflash,format=raw"
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
	}
	if blkdev.Transport.isVirtioCCW(nil) {
		blkdev.DevNo = DevNo
	}
	testAppend(blkdev, deviceBlockString, t)
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
