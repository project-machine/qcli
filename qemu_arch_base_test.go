//go:build !s390x
// +build !s390x

/*
// Copyright contributors to the Virtual Machine Manager for Go project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/

package qcli

var (
	fullVMCommand = "-enable-kvm -smp 4 -m 4096 -cpu qemu64,+x2apic -M q35,smm=on,accel=kvm -global ICH9-LPC.disable_s3=1 -global driver=cfi.pflash01,property=secure,value=on -no-hpet -object rng-random,filename=/dev/urandom,id=rng0 -device virtio-rng-pci,rng=rng0,bus=pcie.0,addr=3 -device pcie-root-port,port=0x0,chassis=0x0,id=root-port.4.0,addr=0x4.0x0,multifunction=on -device pcie-root-port,port=0x1,chassis=0x1,id=root-port.4.1,addr=0x4.0x1 -device pcie-root-port,port=0x2,chassis=0x2,id=root-port.4.2,addr=0x4.0x2 -device pcie-root-port,port=0x3,chassis=0x3,id=root-port.4.3,addr=0x4.0x3 -device pcie-root-port,port=0x4,chassis=0x4,id=root-port.4.4,addr=0x4.0x4 -device pcie-root-port,port=0x5,chassis=0x5,id=root-port.4.5,addr=0x4.0x5 -device pcie-root-port,port=0x6,chassis=0x6,id=root-port.4.6,addr=0x4.0x6 -device pcie-root-port,port=0x7,chassis=0x7,id=root-port.4.7,addr=0x4.0x7 -snapshot -nographic -drive if=pflash,format=raw,readonly,file=/usr/share/OVMF/OVMF_CODE.fd -drive if=pflash,format=raw,file=uefi-nvram.fd -serial mon:stdio -monitor unix:/tmp/vmsockets-1929868109/monitor.socket,server,nowait -mem-path /dev/hugepages -drive file=barehost-lvm-uefi.qcow2,id=drive0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device virtio-blk,drive=drive0,serial=ssd-barehost-lvm-uefi,bus=pcie.0,bootindex=0,logical_block_size=512,physical_block_size=512 -device virtio-net,mac=52:54:00:a2:34:02,netdev=nic1,bus=pcie.0 -netdev user,id=nic1"
)
