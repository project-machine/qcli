package qemu

import "testing"

var (
	deviceSCSIControllerStr        = "-device virtio-scsi-pci,id=foo,disable-modern=false,romfile=efi-virtio.rom"
	deviceSCSIControllerBusAddrStr = "-device virtio-scsi-pci,id=foo,bus=pci.0,addr=00:04.0,disable-modern=true,iothread=iothread1,romfile=efi-virtio.rom"
)

func TestAppendDeviceSCSIController(t *testing.T) {
	scsiCon := SCSIController{
		ID:      "foo",
		ROMFile: romfile,
	}

	if scsiCon.Transport.isVirtioCCW(nil) {
		scsiCon.DevNo = DevNo
	}

	testAppend(scsiCon, deviceSCSIControllerStr, t)

	scsiCon.Bus = "pci.0"
	scsiCon.Addr = "00:04.0"
	scsiCon.DisableModern = true
	scsiCon.IOThread = "iothread1"
	testAppend(scsiCon, deviceSCSIControllerBusAddrStr, t)
}
