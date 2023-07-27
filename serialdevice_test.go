package qcli

import "testing"

var (
	deviceLegacySerialMonMuxString = "-serial mon:stdio"
	deviceLegacySerialString       = "-serial chardev:serial0"
	deviceLegacySerialSocketString = "-serial unix:/tmp/serial.sock,server=on,wait=off"
	deviceSerialString             = "-device virtio-serial-pci,id=serial0,romfile=efi-virtio.rom,disable-modern=true,max_ports=2"
	deviceVirtioSerialPortString   = "-device virtserialport,chardev=char0,id=channel0,name=channel.0 -chardev socket,id=char0,path=/tmp/char.sock,server=on,wait=off"
	deviceSpiceSerialPortString    = "-device virtserialport,chardev=spicechannel0,name=com.redhat.spice.0 -chardev spicevmc,id=spicechannel0,name=vdagent"
	devicePCISerialDeviceString    = "-device pci-serial,id=pciser0,chardev=serial0"
	devicePCISerialDeviceString2x1 = "-device pci-serial-2x,id=pciser0,chardev1=serial0"
	devicePCISerialDeviceString2x2 = "-device pci-serial-2x,id=pciser0,chardev1=serial0,chardev2=serial1"
	devicePCISerialDeviceString4x2 = "-device pci-serial-4x,id=pciser0,chardev1=serial0,chardev2=serial1"
	devicePCISerialDeviceString4x4 = "-device pci-serial-4x,id=pciser0,multifunction=on,chardev1=serial0,chardev2=serial1,chardev3=serial2,chardev4=serial3"
)

func TestAppendLegacySerialMonMux(t *testing.T) {
	sdev := LegacySerialDevice{
		MonMux: true,
	}

	testAppend(sdev, deviceLegacySerialMonMuxString, t)
}

func TestAppendLegacySerial(t *testing.T) {
	sdev := LegacySerialDevice{
		ChardevID: "serial0",
	}

	testAppend(sdev, deviceLegacySerialString, t)
}

func TestAppendLegacySerialUnix(t *testing.T) {
	mon := LegacySerialDevice{
		Backend: Socket,
		Path:    "/tmp/serial.sock",
	}
	testAppend(mon, deviceLegacySerialSocketString, t)

}

func TestAppendDeviceVirtSerial(t *testing.T) {
	sdev := SerialDevice{
		Driver:        VirtioSerial,
		ID:            "serial0",
		DisableModern: true,
		ROMFile:       romfile,
		MaxPorts:      2,
	}
	if sdev.Transport.isVirtioCCW(nil) {
		sdev.DevNo = DevNo
	}

	testAppend(sdev, deviceSerialString, t)
}

func TestAppendDeviceSerialPort(t *testing.T) {
	chardev := CharDevice{
		Driver:   VirtioSerialPort,
		Backend:  Socket,
		ID:       "char0",
		DeviceID: "channel0",
		Path:     "/tmp/char.sock",
		Name:     "channel.0",
	}
	if chardev.Transport.isVirtioCCW(nil) {
		chardev.DevNo = DevNo
	}
	testAppend(chardev, deviceVirtioSerialPortString, t)
}

func TestAppendVirtioDeviceSerialPort(t *testing.T) {
	chardev := CharDevice{
		Driver:   VirtioSerialPort,
		Backend:  Socket,
		ID:       "char0",
		DeviceID: "channel0",
		Path:     "/tmp/char.sock",
		Name:     "channel.0",
	}
	if chardev.Transport.isVirtioCCW(nil) {
		chardev.DevNo = DevNo
	}
	testAppend(chardev, deviceVirtioSerialPortString, t)
}

func TestAppendEmptySerialDevice(t *testing.T) {
	device := SerialDevice{}

	if err := device.Valid(); err == nil {
		t.Fatalf("SerialDevice should not be valid when ID is empty")
	}
}

func TestAppendDevicePCISerialDevice(t *testing.T) {
	sdev := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0"},
		MaxPorts:   1,
	}
	testAppend(sdev, devicePCISerialDeviceString, t)
}

func TestAppendDevicePCISerialDevice2x1Char(t *testing.T) {
	sdev := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0"},
		MaxPorts:   2,
	}
	testAppend(sdev, devicePCISerialDeviceString2x1, t)
}

func TestAppendDevicePCISerialDevice2x2Char(t *testing.T) {
	sdev := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0", "serial1"},
		MaxPorts:   2,
	}
	testAppend(sdev, devicePCISerialDeviceString2x2, t)
}

func TestAppendDevicePCISerialDevice4x2Char(t *testing.T) {
	sdev := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0", "serial1"},
		MaxPorts:   4,
	}
	testAppend(sdev, devicePCISerialDeviceString4x2, t)
}

func TestAppendDevicePCISerialDevice4x4Char(t *testing.T) {
	sdev := SerialDevice{
		Driver:        PCISerialDevice,
		ID:            "pciser0",
		ChardevIDs:    []string{"serial0", "serial1", "serial2", "serial3"},
		MaxPorts:      4,
		Multifunction: true,
	}
	testAppend(sdev, devicePCISerialDeviceString4x4, t)
}

func TestAppendMalformedPCISerialDeviceChardevIDs(t *testing.T) {
	device := SerialDevice{
		Driver:   PCISerialDevice,
		ID:       "pciser0",
		MaxPorts: 2,
	}

	if err := device.Valid(); err == nil {
		t.Fatalf("SerialDevice should not have empty ChardevIDs list")
	}
}

func TestAppendLongPCISerialDeviceChardevIDs(t *testing.T) {
	device := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0", "serial1", "serial2", "serial3", "serial4"},
		MaxPorts:   2,
	}

	if err := device.Valid(); err == nil {
		t.Fatalf("SerialDevice should not have ChardevIDs list of length > 4")
	}
}
