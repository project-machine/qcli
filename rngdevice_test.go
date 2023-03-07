package qcli

import (
	"fmt"
	"testing"
)

func TestAppendVirtioRng(t *testing.T) {
	var objectString = "-object rng-random,id=rng0"
	var deviceString = "-device " + string(VirtioRng)

	rngDevice := RngDevice{
		ID:      "rng0",
		Driver:  VirtioRng,
		ROMFile: romfile,
		Addr:    "3",
	}

	deviceString += "-" + rngDevice.Transport.getName(nil) + ",rng=rng0,addr=0x03"
	if romfile != "" {
		deviceString = deviceString + ",romfile=efi-virtio.rom"
	}

	if rngDevice.Transport.isVirtioCCW(nil) {
		rngDevice.DevNo = DevNo
		deviceString += ",devno=" + rngDevice.DevNo
	}

	testAppend(rngDevice, objectString+" "+deviceString, t)

	rngDevice.Filename = "/dev/urandom"
	objectString += ",filename=" + rngDevice.Filename

	testAppend(rngDevice, objectString+" "+deviceString, t)

	rngDevice.MaxBytes = 20

	deviceString += fmt.Sprintf(",max-bytes=%d", rngDevice.MaxBytes)
	testAppend(rngDevice, objectString+" "+deviceString, t)

	rngDevice.Period = 500

	deviceString += fmt.Sprintf(",period=%d", rngDevice.Period)
	testAppend(rngDevice, objectString+" "+deviceString, t)

}

func TestVirtioRngValid(t *testing.T) {
	rng := RngDevice{}

	if err := rng.Valid(); err == nil {
		t.Fatalf("rng should not be valid when ID is empty")
	}

	rng.ID = "rng0"
	if err := rng.Valid(); err == nil {
		t.Fatalf("rng should not be valid when Driver is empty")
	}

	rng.Driver = VirtioRng
	if err := rng.Valid(); err != nil {
		t.Fatalf("rng should be valid")
	}
}

func TestAppendVirtioRngPCIEBusAddr(t *testing.T) {
	deviceRngPCIeBusAddr := "-object rng-random,id=rng0,filename=/dev/urandom -device virtio-rng-pci,rng=rng0,bus=pcie.0,addr=0x03"

	rngDevice := RngDevice{
		Driver:    VirtioRng,
		ID:        "rng0",
		Bus:       "pcie.0",
		Addr:      "3",
		Transport: TransportPCI,
		Filename:  RngDevUrandom,
	}

	testAppend(rngDevice, deviceRngPCIeBusAddr, t)
}
