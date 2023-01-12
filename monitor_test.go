package qcli

import "testing"

var (
	deviceMonitorString          = "-monitor chardev:char0"
	deviceMonitorSerialMuxString = "-monitor chardev:char0 -serial chardev:char0 -chardev stdio,id=char0,mux=on,signal=off"
	deviceMonitorSocketString    = "-monitor unix:/tmp/mon.sock,server=on,wait=off"
)

func TestAppendMonitor(t *testing.T) {
	mon := MonitorDevice{
		ChardevID: "char0",
	}

	testAppend(mon, deviceMonitorString, t)
}

func TestAppendMonitorSerialMux(t *testing.T) {
	cdev := CharDevice{
		Driver:  LegacySerial,
		Backend: Stdio,
		ID:      "char0",
		Mux:     "on",
		Signal:  "off",
	}

	serial := LegacySerialDevice{
		ChardevID: "char0",
	}

	mon := MonitorDevice{
		ChardevID: "char0",
	}

	c := &Config{}
	c.devices = []Device{mon, serial, cdev}

	testConfig(c, deviceMonitorSerialMuxString, t)
}

func TestAppendMonitorSocket(t *testing.T) {
	mon := MonitorDevice{
		Backend: Socket,
		Path:    "/tmp/mon.sock",
	}
	testAppend(mon, deviceMonitorSocketString, t)
}
