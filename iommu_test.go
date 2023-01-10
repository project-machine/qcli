package qemu

import "testing"

var (
	vIommuString        = "-device intel-iommu,intremap=on,device-iotlb=on,caching-mode=on"
	vIommuNoCacheString = "-device intel-iommu,intremap=on,device-iotlb=on,caching-mode=off"
)

func TestIommu(t *testing.T) {
	iommu := IommuDev{
		Intremap:    true,
		DeviceIotlb: true,
		CachingMode: true,
	}

	if err := iommu.Valid(); err != nil {
		t.Fatalf("iommu should be valid")
	}

	testAppend(iommu, vIommuString, t)

	iommu.CachingMode = false

	testAppend(iommu, vIommuNoCacheString, t)

}
