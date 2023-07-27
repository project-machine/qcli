package qcli

import "testing"

var (
	deviceCharDeviceBackendFile     = "-chardev file,id=serial0,path=/tmp/serial.log"
	deviceCharDeviceBackendSocket   = "-chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off"
	deviceCharDeviceBackendStdioMux = "-chardev stdio,id=serial0,mux=on,signal=off"
	deviceCharDeviceMultiple        = "-chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -chardev socket,id=monitor0,path=/tmp/monitor.sock,server=on,wait=off"
	deviceCharDevicePCIDriver       = "-serial none -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -device pci-serial,id=pciser0,chardev=serial0"
	deviceCharDevicePCIDriver2x     = "-serial none -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -device pci-serial-2x,id=pciser0,chardev1=serial0"
)

func TestBadCharDevice(t *testing.T) {
	c := &Config{
		CharDevices: []CharDevice{
			CharDevice{},
			CharDevice{
				ID:   "id1",
				Path: "",
			},
			CharDevice{
				ID:   "",
				Path: "/tmp/foo",
			},
		},
	}
	c.appendDevices()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams for BadCharDevices, found %s", c.qemuParams)
	}
}

func TestAppendCharDeviceFile(t *testing.T) {
	chardev := CharDevice{
		Driver:  LegacySerial,
		Backend: File,
		ID:      "serial0",
		Path:    "/tmp/serial.log",
	}

	testAppend(chardev, deviceCharDeviceBackendFile, t)
}

func TestAppendCharDeviceBackendStdioMux(t *testing.T) {
	chardev := CharDevice{
		Driver:  LegacySerial,
		Backend: Stdio,
		ID:      "serial0",
		Mux:     "on",
		Signal:  "off",
	}
	testAppend(chardev, deviceCharDeviceBackendStdioMux, t)
}

func TestAppendCharDeviceBackendSocket(t *testing.T) {
	chardev := CharDevice{
		Driver:  LegacySerial,
		Backend: Socket,
		ID:      "serial0",
		Path:    "/tmp/console.sock",
	}

	testAppend(chardev, deviceCharDeviceBackendSocket, t)
}

func TestAppendMultipleCharDevices(t *testing.T) {
	c := &Config{}
	serial := CharDevice{
		Driver:  LegacySerial,
		Backend: Socket,
		ID:      "serial0",
		Path:    "/tmp/console.sock",
	}
	mon := CharDevice{
		Driver:  LegacySerial,
		Backend: Socket,
		ID:      "monitor0",
		Path:    "/tmp/monitor.sock",
	}
	c.CharDevices = []CharDevice{serial, mon}
	testConfig(c, deviceCharDeviceMultiple, t)
}

func TestAppendPCIDriver1x(t *testing.T) {
	c := &Config{}
	serial := CharDevice{
		Driver:  PCISerialDevice,
		Backend: Socket,
		ID:      "serial0",
		Path:    "/tmp/console.sock",
	}
	pcidev := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0"},
		MaxPorts:   1,
	}
	c.CharDevices = []CharDevice{serial}
	c.SerialDevices = []SerialDevice{pcidev}
	testConfig(c, deviceCharDevicePCIDriver, t)
}

func TestAppendPCIDriver2x1(t *testing.T) {
	c := &Config{}
	serial := CharDevice{
		Driver:  PCISerialDevice,
		Backend: Socket,
		ID:      "serial0",
		Path:    "/tmp/console.sock",
	}
	pcidev := SerialDevice{
		Driver:     PCISerialDevice,
		ID:         "pciser0",
		ChardevIDs: []string{"serial0"},
		MaxPorts:   2,
	}
	c.CharDevices = []CharDevice{serial}
	c.SerialDevices = []SerialDevice{pcidev}
	testConfig(c, deviceCharDevicePCIDriver2x, t)
}
