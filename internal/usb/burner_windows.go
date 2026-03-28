//go:build windows

package usb

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32              = windows.NewLazySystemDLL("kernel32.dll")
	procCreateFileW       = kernel32.NewProc("CreateFileW")
	procDeviceIoControl   = kernel32.NewProc("DeviceIoControl")
	procFindFirstVolumeW = kernel32.NewProc("FindFirstVolumeW")
	procFindNextVolumeW  = kernel32.NewProc("FindNextVolumeW")
	procFindVolumeClose  = kernel32.NewProc("FindVolumeClose")
)

const (
	fsctlLockVolume      = 0x00090018
	fsctlDismountVolume  = 0x00090020
	ioctlStorageGetDeviceNumber = 0x002D1080
)

type storageDeviceNumber struct {
	DeviceType      uint32
	DeviceNumber    uint32
	PartitionNumber uint32
}

// openDeviceForWrite locks and dismounts all volumes on the physical drive,
// then opens it with unbuffered I/O for direct sector writes.
func openDeviceForWrite(devicePath string) (*os.File, error) {
	// Extract disk number from path like \\.\PhysicalDrive3
	diskNum, err := parseDiskNumber(devicePath)
	if err != nil {
		return nil, err
	}

	// Lock and dismount all volumes that live on this physical drive
	if err := lockAndDismountVolumes(diskNum); err != nil {
		return nil, fmt.Errorf("failed to dismount volumes: %v", err)
	}

	// Open the physical drive with FILE_FLAG_NO_BUFFERING | FILE_FLAG_WRITE_THROUGH
	pathPtr, _ := syscall.UTF16PtrFromString(devicePath)
	handle, _, callErr := procCreateFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		0,
		syscall.OPEN_EXISTING,
		0x20000000|0x80000000, // FILE_FLAG_NO_BUFFERING | FILE_FLAG_WRITE_THROUGH
		0,
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return nil, fmt.Errorf("failed to open device (try running as Administrator): %v", callErr)
	}

	return os.NewFile(handle, devicePath), nil
}

func parseDiskNumber(path string) (uint32, error) {
	upper := strings.ToUpper(path)
	var num uint32
	// Handle both \\.\PhysicalDrive3 and \\.\PHYSICALDRIVE3
	for _, prefix := range []string{`\\.\PHYSICALDRIVE`, `\\.\PHYSICALDISK`} {
		if strings.HasPrefix(upper, prefix) {
			_, err := fmt.Sscanf(upper[len(prefix):], "%d", &num)
			if err == nil {
				return num, nil
			}
		}
	}
	return 0, fmt.Errorf("cannot parse disk number from: %s", path)
}

func lockAndDismountVolumes(diskNumber uint32) error {
	volumes, err := findVolumesOnDisk(diskNumber)
	if err != nil {
		return err
	}

	for _, volPath := range volumes {
		if err := lockAndDismountVolume(volPath); err != nil {
			// Non-fatal: some volumes may already be dismounted
			continue
		}
	}
	return nil
}

func lockAndDismountVolume(volumePath string) error {
	// Volume paths end with '\', but CreateFile needs it without
	path := strings.TrimRight(volumePath, `\`)
	pathPtr, _ := syscall.UTF16PtrFromString(path)

	handle, _, err := procCreateFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		0,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return fmt.Errorf("open volume %s: %v", path, err)
	}
	// We intentionally do NOT close this handle — it must stay open to keep the lock.
	// It will be released when the process exits.

	var bytesReturned uint32

	// FSCTL_LOCK_VOLUME
	ret, _, err := procDeviceIoControl.Call(
		handle,
		fsctlLockVolume,
		0, 0,
		0, 0,
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if ret == 0 {
		return fmt.Errorf("lock volume %s: %v", path, err)
	}

	// FSCTL_DISMOUNT_VOLUME
	ret, _, err = procDeviceIoControl.Call(
		handle,
		fsctlDismountVolume,
		0, 0,
		0, 0,
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if ret == 0 {
		return fmt.Errorf("dismount volume %s: %v", path, err)
	}

	return nil
}

func findVolumesOnDisk(diskNumber uint32) ([]string, error) {
	var volumes []string

	buf := make([]uint16, 260)
	findHandle, _, err := procFindFirstVolumeW.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if findHandle == uintptr(syscall.InvalidHandle) {
		return nil, fmt.Errorf("FindFirstVolume: %v", err)
	}
	defer procFindVolumeClose.Call(findHandle)

	for {
		volName := syscall.UTF16ToString(buf)
		if volumeIsOnDisk(volName, diskNumber) {
			volumes = append(volumes, volName)
		}

		ret, _, _ := procFindNextVolumeW.Call(
			findHandle,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
		)
		if ret == 0 {
			break
		}
	}

	return volumes, nil
}

func volumeIsOnDisk(volumePath string, diskNumber uint32) bool {
	// Open the volume to query which physical disk it belongs to
	path := strings.TrimRight(volumePath, `\`)
	pathPtr, _ := syscall.UTF16PtrFromString(path)

	handle, _, _ := procCreateFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0, // No access needed for the query
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		0,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return false
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	var sdn storageDeviceNumber
	var bytesReturned uint32

	ret, _, _ := procDeviceIoControl.Call(
		handle,
		ioctlStorageGetDeviceNumber,
		0, 0,
		uintptr(unsafe.Pointer(&sdn)),
		uintptr(unsafe.Sizeof(sdn)),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if ret == 0 {
		return false
	}

	return sdn.DeviceNumber == diskNumber
}
