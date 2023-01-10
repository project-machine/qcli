package qemu

import "testing"

func TestAppendPVPanicDevice(t *testing.T) {
	testCases := []struct {
		dev Device
		out string
	}{
		{nil, ""},
		{PVPanicDevice{}, "-device pvpanic"},
		{PVPanicDevice{NoShutdown: true}, "-device pvpanic -no-shutdown"},
	}

	for _, tc := range testCases {
		testAppend(tc.dev, tc.out, t)
	}
}
