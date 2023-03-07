package qcli

import "testing"

var (
	deviceSCSIControllerStr        = "-device virtio-scsi-pci,id=foo,addr=0x1e,bus=pcie.0,disable-modern=false,romfile=efi-virtio.rom"
	deviceSCSIControllerBusAddrStr = "-device virtio-scsi-pci,id=foo,addr=0x1e,bus=pci.0,disable-modern=true,iothread=iothread1,romfile=efi-virtio.rom -object iothread,poll-max-ns=32,id=iothread1"
)

func TestAppendDeviceSCSIController(t *testing.T) {
	scsiCon := SCSIControllerDevice{
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
