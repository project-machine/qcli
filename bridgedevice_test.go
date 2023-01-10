package qcli

import "testing"

var (
	devicePCIBridgeString         = "-device pci-bridge,bus=/pci-bus/pcie.0,id=mybridge,chassis_nr=5,shpc=on,addr=ff,romfile=efi-virtio.rom"
	devicePCIBridgeStringReserved = "-device pci-bridge,bus=/pci-bus/pcie.0,id=mybridge,chassis_nr=5,shpc=off,addr=ff,romfile=efi-virtio.rom,io-reserve=4k,mem-reserve=1m,pref64-reserve=1m"
	devicePCIEBridgeString        = "-device pcie-pci-bridge,bus=/pci-bus/pcie.0,id=mybridge,addr=ff,romfile=efi-virtio.rom"
	romfile                       = "efi-virtio.rom"
)

func TestAppendPCIBridgeDevice(t *testing.T) {

	bridge := BridgeDevice{
		Type:    PCIBridge,
		ID:      "mybridge",
		Bus:     "/pci-bus/pcie.0",
		Addr:    "255",
		Chassis: 5,
		SHPC:    true,
		ROMFile: romfile,
	}

	testAppend(bridge, devicePCIBridgeString, t)
}

func TestAppendPCIBridgeDeviceWithReservations(t *testing.T) {

	bridge := BridgeDevice{
		Type:          PCIBridge,
		ID:            "mybridge",
		Bus:           "/pci-bus/pcie.0",
		Addr:          "255",
		Chassis:       5,
		SHPC:          false,
		ROMFile:       romfile,
		IOReserve:     "4k",
		MemReserve:    "1m",
		Pref64Reserve: "1m",
	}

	testAppend(bridge, devicePCIBridgeStringReserved, t)
}

func TestAppendPCIEBridgeDevice(t *testing.T) {

	bridge := BridgeDevice{
		Type:    PCIEBridge,
		ID:      "mybridge",
		Bus:     "/pci-bus/pcie.0",
		Addr:    "255",
		ROMFile: "efi-virtio.rom",
	}

	testAppend(bridge, devicePCIEBridgeString, t)
}
