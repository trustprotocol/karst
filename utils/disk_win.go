// +build windows

package utils

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

func NewDiskUsage(path string) (*DiskStatus, error) {
	h := windows.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := uint64(0)
	lpTotalNumberOfBytes := uint64(0)
	lpTotalNumberOfFreeBytes := uint64(0)
	_, _, _ = c.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("C:"))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))

	return &DiskStatus{
		All:  lpTotalNumberOfBytes,
		Free: lpTotalNumberOfFreeBytes,
		Used: lpFreeBytesAvailable,
	}, nil
}
