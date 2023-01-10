package qemu

import "testing"

var (
	deviceVFIOString           = "-device vfio-pci,host=02:10.0,x-pci-vendor-id=0x1234,x-pci-device-id=0x5678,romfile=efi-virtio.rom"
	deviceVFIOPCIeSimpleString = "-device vfio-pci,host=02:00.0,bus=rp0"
	deviceVFIOPCIeFullString   = "-device vfio-pci,host=02:00.0,x-pci-vendor-id=0x10de,x-pci-device-id=0x15f8,romfile=efi-virtio.rom,bus=rp1"
)

func TestAppendDeviceVFIO(t *testing.T) {
	vfioDevice := VFIODevice{
		BDF:      "02:10.0",
		ROMFile:  romfile,
		VendorID: "0x1234",
		DeviceID: "0x5678",
	}

	if vfioDevice.Transport.isVirtioCCW(nil) {
		vfioDevice.DevNo = DevNo
	}

	testAppend(vfioDevice, deviceVFIOString, t)
}

func TestAppendDeviceVFIOPCIe(t *testing.T) {
	// default test
	pcieRootPortID := "rp0"
	vfioDevice := VFIODevice{
		BDF: "02:00.0",
		Bus: pcieRootPortID,
	}
	testAppend(vfioDevice, deviceVFIOPCIeSimpleString, t)

	// full test
	pcieRootPortID = "rp1"
	vfioDevice = VFIODevice{
		BDF:      "02:00.0",
		Bus:      pcieRootPortID,
		ROMFile:  romfile,
		VendorID: "0x10de",
		DeviceID: "0x15f8",
	}
	testAppend(vfioDevice, deviceVFIOPCIeFullString, t)
}
