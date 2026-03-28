//go:build windows

package usb

import "syscall"

func readFirstSector(devicePath string) ([]byte, error) {
	pathPtr, _ := syscall.UTF16PtrFromString(devicePath)

	handle, err := syscall.CreateFile(
		pathPtr,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(handle)

	buf := make([]byte, 512)
	var bytesRead uint32
	err = syscall.ReadFile(handle, buf, &bytesRead, nil)
	if err != nil {
		return nil, err
	}

	return buf[:bytesRead], nil
}
