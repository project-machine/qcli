package qcli

import "testing"

func TestSpiceDevice(t *testing.T) {
	testCases := []struct {
		dev Device
		out string
	}{
		{SpiceDevice{Port: "5901"}, "-spice port=5901,addr=127.0.0.1 -device virtio-serial-pci -device virtserialport,chardev=spicechannel0,name=com.redhat.spice.0 -chardev spicevmc,id=spicechannel0,name=vdagent"},
		{SpiceDevice{TLSPort: "5902", HostAddress: "0.0.0.0", DisableTicketing: true}, "-spice tls-port=5902,addr=0.0.0.0,disable-ticketing=on -device virtio-serial-pci -device virtserialport,chardev=spicechannel0,name=com.redhat.spice.0 -chardev spicevmc,id=spicechannel0,name=vdagent"},
	}

	for _, tc := range testCases {
		testAppend(tc.dev, tc.out, t)
	}
}

func TestSpiceDeviceInvalid(t *testing.T) {
	dev := SpiceDevice{}

	if err := dev.Valid(); err == nil {
		t.Fatalf("A SpiceDevice with missing Port and TLSPort fields is NOT valid")
	}

	dev.Port = "5901"
	dev.TLSPort = "5902"

	if err := dev.Valid(); err == nil {
		t.Fatalf("A SpiceDevice with both Port and TLSPort fields is NOT valid")
	}
}
