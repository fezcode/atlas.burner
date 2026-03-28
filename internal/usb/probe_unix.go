//go:build !windows

package usb

import "os"

func readFirstSector(devicePath string) ([]byte, error) {
	f, err := os.Open(devicePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)
	return buf, err
}
