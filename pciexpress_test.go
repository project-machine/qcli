package qemu

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var (
	devicePCIeRootPortSimpleString = "-device pcie-root-port,id=rp1,bus=pcie.0,chassis=0x00,slot=0x00,addr=0x00,multifunction=off"
	devicePCIeRootPortFullString   = "-device pcie-root-port,id=rp2,bus=pcie.0,chassis=0x0,slot=0x1,addr=0x2,multifunction=on,bus-reserve=0x3,pref64-reserve=16G,mem-reserve=1G,io-reserve=512M,romfile=efi-virtio.rom"
)

func TestDevicePCIeRootPortValid(t *testing.T) {
	pcieRootPortDevice := PCIeRootPortDevice{}
	if err := pcieRootPortDevice.Valid(); err == nil {
		t.Fatalf("PCIeRootPort should NOT be valid when ID is empty")
	}

	pcieRootPortDevice.ID = "rp0"
	if err := pcieRootPortDevice.Valid(); err != nil {
		t.Fatalf("PCIeRootPort should be valid")
	}

	pcieRootPortDevice.Pref32Reserve = "256M"
	pcieRootPortDevice.Pref64Reserve = "16G"
	if err := pcieRootPortDevice.Valid(); err == nil {
		t.Fatalf("PCIeRootPort should NOT be valid, Pref32Reserve and Pref64Reserve are mutually exclusive")
	}
}

func TestAppendDevicePCIeRootPortSimple(t *testing.T) {
	pcieRootPortDevice := PCIeRootPortDevice{
		ID: "rp1",
	}
	testAppend(pcieRootPortDevice, devicePCIeRootPortSimpleString, t)

}

func TestAppendDevicePCIeRootPortFull(t *testing.T) {
	pcieRootPortDevice := PCIeRootPortDevice{
		ID:            "rp2",
		Multifunction: true,
		Bus:           "pcie.0",
		Chassis:       "0x0",
		Slot:          "0x1",
		Addr:          "0x2",
		Pref64Reserve: "16G",
		IOReserve:     "512M",
		MemReserve:    "1G",
		BusReserve:    "0x3",
		ROMFile:       romfile,
	}
	testAppend(pcieRootPortDevice, devicePCIeRootPortFullString, t)
}

func TestAppendDevicePCIeRootPortMultiFuncPair(t *testing.T) {
	c := &Config{
		PCIeRootPortDevices: []PCIeRootPortDevice{
			PCIeRootPortDevice{
				ID:            "root-port.4.0",
				Bus:           "pcie.0",
				Chassis:       "0x0",
				Slot:          "0x00",
				Port:          "0x0",
				Addr:          "0x4.0x0",
				Multifunction: true,
			},
			PCIeRootPortDevice{
				ID:            "root-port.4.1",
				Bus:           "pcie.0",
				Chassis:       "0x1",
				Slot:          "0x00",
				Port:          "0x1",
				Addr:          "0x4.0x1",
				Multifunction: false,
			},
		},
	}
	expected := "-device pcie-root-port,id=root-port.4.0,bus=pcie.0,chassis=0x0,slot=0x00,port=0x0,addr=0x4.0x0,multifunction=on -device pcie-root-port,id=root-port.4.1,bus=pcie.0,chassis=0x1,slot=0x00,port=0x1,addr=0x4.0x1"
	testConfig(c, expected, t)
}

func TestAppendDevicePCIeRootMultifunctionPortRange(t *testing.T) {
	portPrefix := "root-port"
	bus := "pcie.0"
	baseAddr := "4"
	numPorts := 8
	devices := []Device{}
	expectedParams := []string{}

	// hand generate devices and expected results
	for p := 0; p < numPorts; p++ {
		rootPortID := fmt.Sprintf("%s.%s.%d", portPrefix, baseAddr, p)
		port := fmt.Sprintf("0x%x", p)
		chassis := fmt.Sprintf("0x%x", p)
		addr := fmt.Sprintf("%s.0x%x", baseAddr, p)
		expected := fmt.Sprintf("-device pcie-root-port,id=%s,bus=%s,chassis=%s,slot=0x00,port=%s,addr=%s", rootPortID, bus, chassis, port, addr)

		pcieRootPort := PCIeRootPortDevice{
			ID:      rootPortID,
			Port:    port,
			Chassis: chassis,
			Addr:    addr,
			Bus:     bus,
		}

		if p == 0 {
			pcieRootPort.Multifunction = true
			expected = expected + ",multifunction=on"
		}

		// verify we got the string we wanted
		var config Config
		testConfigAppend(&config, pcieRootPort, expected, t)

		// Add this to the list of devices
		expectedParams = append(expectedParams, expected)
		devices = append(devices, pcieRootPort)
	}

	// test them all together
	config := &Config{devices: devices}
	testConfig(config, strings.Join(expectedParams, " "), t)

	// Use NewPCIeRootPortRange() and compare to current devices
	newDevices, err := NewPCIeRootMultifunctionPortRange(portPrefix, bus, baseAddr, numPorts)
	if err != nil {
		t.Errorf("NewPCIeRootMultifunctionPortRoage returned error: %s", err.Error())
	}

	ok := reflect.DeepEqual(devices, newDevices)
	if !ok {
		t.Errorf("PCIeRootMultifunctionPortRage mismatch, expected %+v, found %+v", devices, newDevices)
	}
}
