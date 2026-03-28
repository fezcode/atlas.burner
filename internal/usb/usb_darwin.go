//go:build darwin

package usb

func GetRemovableDevices() ([]Device, error) {
	// Mock for macOS for now
	return []Device{}, nil
}
