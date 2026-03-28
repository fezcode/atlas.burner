//go:build !windows

package usb

import (
	"fmt"
	"os"
)

// openDeviceForWrite opens a block device for direct writing on Unix.
func openDeviceForWrite(devicePath string) (*os.File, error) {
	f, err := os.OpenFile(devicePath, os.O_WRONLY|os.O_SYNC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open device (try running as root): %v", err)
	}
	return f, nil
}
