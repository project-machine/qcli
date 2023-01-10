package qemu

import "testing"

var (
	deviceVSOCKString = "-device vhost-vsock-pci,disable-modern=true,id=vhost-vsock-pci0,guest-cid=4,romfile=efi-virtio.rom"
)

func TestAppendVSOCK(t *testing.T) {
	vsockDevice := VSOCKDevice{
		ID:            "vhost-vsock-pci0",
		ContextID:     4,
		VHostFD:       nil,
		DisableModern: true,
		ROMFile:       romfile,
	}

	if vsockDevice.Transport.isVirtioCCW(nil) {
		vsockDevice.DevNo = DevNo
	}

	testAppend(vsockDevice, deviceVSOCKString, t)
}

func TestVSOCKValid(t *testing.T) {
	vsockDevice := VSOCKDevice{
		ID:            "vhost-vsock-pci0",
		ContextID:     MinimalGuestCID - 1,
		VHostFD:       nil,
		DisableModern: true,
	}

	if err := vsockDevice.Valid(); err == nil {
		t.Fatalf("VSOCK Context ID is not valid")
	}

	vsockDevice.ContextID = MaxGuestCID + 1

	if err := vsockDevice.Valid(); err == nil {
		t.Fatalf("VSOCK Context ID is not valid")
	}

	vsockDevice.ID = ""

	if err := vsockDevice.Valid(); err == nil {
		t.Fatalf("VSOCK ID is not valid")
	}
}
