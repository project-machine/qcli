package qcli

import "testing"

var (
	deviceIDEControllerPIIX3Str       = "-device piix3-ide,id=ide0,addr=0x1e,bus=pcie.0"
	deviceIDEControllerPIIX4Str       = "-device piix4-ide,id=ide0,addr=0x1e,bus=pcie.0"
	deviceIDEControllerAHCIStr        = "-device ich9-ahci,id=ide0,addr=0x1e,bus=pcie.0"
	deviceIDEControllerAHCIBusAddrStr = "-device ich9-ahci,id=ide0,addr=0x1e,bus=pcie.1,romfile=romfile,rombar=1024,multifunction=on"
)

func TestAppendDeviceIDEController(t *testing.T) {
	ideCon := IDEControllerDevice{
		ID:     "ide0",
		Driver: ICH9AHCIController,
	}
	testAppend(ideCon, deviceIDEControllerAHCIStr, t)

	ideCon.Driver = PIIX3IDEController
	testAppend(ideCon, deviceIDEControllerPIIX3Str, t)

	ideCon.Driver = PIIX4IDEController
	testAppend(ideCon, deviceIDEControllerPIIX4Str, t)

	ideCon.Driver = ICH9AHCIController
	ideCon.Bus = "pcie.1"
	ideCon.Addr = "0x5"
	ideCon.ROMFile = "romfile"
	ideCon.ROMBar = "1024"
	ideCon.Multifunction = true
	testAppend(ideCon, deviceIDEControllerAHCIBusAddrStr, t)
}
