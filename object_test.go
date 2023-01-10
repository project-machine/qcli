package qcli

import "testing"

var (
	memPathString      = "-mem-path /dev/hugepages/vm1 -mem-prealloc"
	deviceNVDIMMString = "-device nvdimm,id=nv0,memdev=mem0,unarmed=on -object memory-backend-file,id=mem0,mem-path=/root,size=65536,readonly=on"
	objectEPCString    = "-object memory-backend-epc,id=epc0,size=65536,prealloc=on"
)

func TestAppendObjectLegacy(t *testing.T) {
	object := Object{
		Type:     LegacyMemPath,
		MemPath:  "/dev/hugepages/vm1",
		Prealloc: true,
	}

	testAppend(object, memPathString, t)
}

func TestAppendObjectDeviceNVDIMM(t *testing.T) {
	object := Object{
		Driver:   NVDIMM,
		Type:     MemoryBackendFile,
		DeviceID: "nv0",
		ID:       "mem0",
		MemPath:  "/root",
		Size:     1 << 16,
		ReadOnly: true,
	}

	testAppend(object, deviceNVDIMMString, t)
}

func TestAppendObjectEPC(t *testing.T) {
	object := Object{
		Type:     MemoryBackendEPC,
		ID:       "epc0",
		Size:     1 << 16,
		Prealloc: true,
	}

	testAppend(object, objectEPCString, t)
}
