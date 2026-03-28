//go:build windows

package usb

import (
	"fmt"
	"strings"

	"github.com/yusufpapurcu/wmi"
)

type Win32_DiskDrive struct {
	DeviceID   string
	Caption    string
	Size       uint64
	MediaType  string
	Partitions uint32
}

func GetRemovableDevices() ([]Device, error) {
	var drives []Win32_DiskDrive
	q := "SELECT DeviceID, Caption, Size, MediaType, Partitions FROM Win32_DiskDrive"
	err := wmi.Query(q, &drives)
	if err != nil {
		return nil, err
	}

	var devices []Device
	for _, drive := range drives {
		if strings.Contains(strings.ToLower(drive.MediaType), "removable") {
			dev := Device{
				Name:        fmt.Sprintf("Drive %s", drive.DeviceID),
				Description: drive.Caption,
				Size:        drive.Size,
				Path:        drive.DeviceID,
				Bootable:    DetectBootType(drive.DeviceID),
			}
			devices = append(devices, dev)
		}
	}
	return devices, nil
}
