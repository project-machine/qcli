package qemu

import "testing"

var (
	qmpSingleSocketServerString = "-qmp unix:cc-qmp,server=on,wait=off"
	qmpSingleSocketString       = "-qmp unix:cc-qmp"
	qmpSocketServerString       = "-qmp unix:cc-qmp-1,server=on,wait=off -qmp unix:cc-qmp-2,server=on,wait=off"
)

func TestAppendSingleQMPSocketServer(t *testing.T) {
	qmp := QMPSocket{
		Type:   "unix",
		Name:   "cc-qmp",
		Server: true,
		NoWait: true,
	}

	testAppend(qmp, qmpSingleSocketServerString, t)
}

func TestAppendSingleQMPSocket(t *testing.T) {
	qmp := QMPSocket{
		Type:   Unix,
		Name:   "cc-qmp",
		Server: false,
	}

	testAppend(qmp, qmpSingleSocketString, t)
}

func TestAppendQMPSocketServer(t *testing.T) {
	qmp := []QMPSocket{
		{
			Type:   "unix",
			Name:   "cc-qmp-1",
			Server: true,
			NoWait: true,
		},
		{
			Type:   "unix",
			Name:   "cc-qmp-2",
			Server: true,
			NoWait: true,
		},
	}

	testAppend(qmp, qmpSocketServerString, t)
}

func TestBadQMPSockets(t *testing.T) {
	c := &Config{}
	c.appendQMPSockets()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		QMPSockets: []QMPSocket{{}},
	}

	c.appendQMPSockets()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		QMPSockets: []QMPSocket{{Name: "test"}},
	}

	c.appendQMPSockets()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		QMPSockets: []QMPSocket{
			{
				Name: "test",
				Type: QMPSocketType("ip"),
			},
		},
	}

	c.appendQMPSockets()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}
