package qemu

import "testing"

var (
	deviceFSString = "-device virtio-9p-pci,disable-modern=true,fsdev=workload9p,mount_tag=rootfs,romfile=efi-virtio.rom -fsdev local,id=workload9p,path=/var/lib/docker/devicemapper/mnt/e31ebda2,security_model=none,multidevs=remap"
)

func TestAppendDeviceFS(t *testing.T) {
	fsdev := FSDevice{
		Driver:        Virtio9P,
		FSDriver:      Local,
		ID:            "workload9p",
		Path:          "/var/lib/docker/devicemapper/mnt/e31ebda2",
		MountTag:      "rootfs",
		SecurityModel: None,
		DisableModern: true,
		ROMFile:       "efi-virtio.rom",
		Multidev:      Remap,
	}

	if fsdev.Transport.isVirtioCCW(nil) {
		fsdev.DevNo = DevNo
	}

	testAppend(fsdev, deviceFSString, t)
}
