package qcli

import "testing"

var (
	deviceMonitorString          = "-monitor chardev:char0"
	deviceMonitorSerialMuxString = "-monitor chardev:char0 -serial chardev:char0 -chardev stdio,id=char0,mux=on,signal=off"
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
