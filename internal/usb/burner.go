package usb

import (
	"fmt"
	"io"
	"os"
)

// BurnImage writes the contents of imagePath directly to devicePath.
// Platform-specific setup (volume locking, dismounting) is handled by
// openDeviceForWrite, which each platform implements.
func BurnImage(imagePath string, devicePath string, progressCallback func(float64)) error {
	in, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}
	defer in.Close()

	stat, err := in.Stat()
	if err != nil {
		return err
	}
	totalSize := stat.Size()

	dev, err := openDeviceForWrite(devicePath)
	if err != nil {
		return err
	}
	defer dev.Close()

	buf := make([]byte, 4*1024*1024) // 4MB buffer (sector-aligned)
	var written int64

	for {
		n, readErr := in.Read(buf)
		if n > 0 {
			// Pad to sector alignment if this is the last chunk
			writeLen := n
			if readErr == io.EOF {
				remainder := n % 512
				if remainder != 0 {
					writeLen = n + (512 - remainder)
					// buf is already large enough and zero-initialized beyond n
					for i := n; i < writeLen; i++ {
						buf[i] = 0
					}
				}
			}

			nw, writeErr := dev.Write(buf[:writeLen])
			if writeErr != nil {
				return fmt.Errorf("write error: %v", writeErr)
			}
			written += int64(nw)
			if progressCallback != nil {
				progressCallback(float64(written) / float64(totalSize))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}
