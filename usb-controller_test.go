package qcli

import "testing"

var (
	deviceUSBControllerQemuXHCIStr        = "-device qemu-xhci,id=usb0,addr=0x1e"
	deviceUSBControllerQemuXHCIBusAddrStr = "-device qemu-xhci,id=usb0,addr=0x1e,romfile=romfile,rombar=1024,multifunction=on"
)

func TestAppendDeviceUSBController(t *testing.T) {
	usbCon := USBControllerDevice{
		ID:     "usb0",
		Driver: USBXHCIController,
	}
	testAppend(usbCon, deviceUSBControllerQemuXHCIStr, t)

	usbCon.Addr = "0x5"
	usbCon.ROMFile = "romfile"
	usbCon.ROMBar = "1024"
	usbCon.Multifunction = true
	testAppend(usbCon, deviceUSBControllerQemuXHCIBusAddrStr, t)
}

func TestAppendDeviceUSBControllerAndUSBCDROM(t *testing.T) {
	conf := &Config{
		USBControllerDevices: []USBControllerDevice{
			USBControllerDevice{
				ID:     "usb0",
				Driver: USBXHCIController,
			},
		},
		BlkDevices: []BlockDevice{
			BlockDevice{
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
			},
		},
	}
	expected := deviceUSBControllerQemuXHCIStr + " " + deviceBlockUSBHDStr
	testConfig(conf, expected, t)
}
