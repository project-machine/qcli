package qcli

import "testing"

var (
	deviceVhostUserNetString  = "-chardev socket,id=char1,path=/tmp/nonexistentsocket.socket -netdev type=vhost-user,id=net1,chardev=char1,vhostforce -device virtio-net-pci,netdev=net1,mac=00:11:22:33:44:55,romfile=efi-virtio.rom"
	deviceVhostUserSCSIString = "-chardev socket,id=char1,path=/tmp/nonexistentsocket.socket -device vhost-user-scsi-pci,id=scsi1,chardev=char1,romfile=efi-virtio.rom"
	deviceVhostUserBlkString  = "-chardev socket,id=char2,path=/tmp/nonexistentsocket.socket -device vhost-user-blk-pci,logical_block_size=4096,size=512M,chardev=char2,romfile=efi-virtio.rom"
)

func TestAppendDeviceVhostUser(t *testing.T) {

	vhostuserBlkDevice := VhostUserDevice{
		SocketPath:    "/tmp/nonexistentsocket.socket",
		CharDevID:     "char2",
		TypeDevID:     "",
		Address:       "",
		VhostUserType: VhostUserBlk,
		ROMFile:       romfile,
	}
	testAppend(vhostuserBlkDevice, deviceVhostUserBlkString, t)

	vhostuserSCSIDevice := VhostUserDevice{
		SocketPath:    "/tmp/nonexistentsocket.socket",
		CharDevID:     "char1",
		TypeDevID:     "scsi1",
		Address:       "",
		VhostUserType: VhostUserSCSI,
		ROMFile:       romfile,
	}
	testAppend(vhostuserSCSIDevice, deviceVhostUserSCSIString, t)

	vhostuserNetDevice := VhostUserDevice{
		SocketPath:    "/tmp/nonexistentsocket.socket",
		CharDevID:     "char1",
		TypeDevID:     "net1",
		Address:       "00:11:22:33:44:55",
		VhostUserType: VhostUserNet,
		ROMFile:       romfile,
	}
	testAppend(vhostuserNetDevice, deviceVhostUserNetString, t)
}
