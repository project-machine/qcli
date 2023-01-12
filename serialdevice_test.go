package qcli

import "testing"

var (
	deviceLegacySerialMonMuxString = "-serial mon:stdio"
	deviceLegacySerialString       = "-serial chardev:serial0"
	deviceLegacySerialSocketString = "-serial unix:/tmp/serial.sock,server=on,wait=off"
	deviceSerialString             = "-device virtio-serial-pci,disable-modern=true,id=serial0,romfile=efi-virtio.rom,max_ports=2"
	deviceVirtioSerialPortString   = "-device virtserialport,chardev=char0,id=channel0,name=channel.0 -chardev socket,id=char0,path=/tmp/char.sock,server=on,wait=off"
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

func TestAppendDeviceSerial(t *testing.T) {
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
