package qemu

import "testing"

func TestLoaderDevice(t *testing.T) {
	testCases := []struct {
		dev Device
		out string
	}{
		{LoaderDevice{File: "f", ID: "id"}, "-device loader,file=f,id=id"},
	}

	for _, tc := range testCases {
		testAppend(tc.dev, tc.out, t)
	}
}

func TestLoaderDeviceInvalid(t *testing.T) {
	dev := LoaderDevice{File: "f", ID: ""}

	if err := dev.Valid(); err == nil {
		t.Fatalf("A LoaderDevice with empty ID field is NOT valid")
	}

	dev.ID = "id"
	dev.File = ""

	if err := dev.Valid(); err == nil {
		t.Fatalf("A LoaderDevice with empty Field field is NOT valid")
	}
}
