package usb

import "fmt"

type BootType string

const (
	BootNone    BootType = ""
	BootMBR     BootType = "MBR"
	BootGPT     BootType = "GPT/UEFI"
	BootUnknown BootType = "Unknown"
)

type Device struct {
	Name        string
	Description string
	Size        uint64
	Path        string // e.g., \\.\PhysicalDrive1 or /dev/sdb
	Bootable    BootType
}

func (d Device) String() string {
	sizeGB := float64(d.Size) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%s (%s) - %.2f GB", d.Name, d.Description, sizeGB)
}

func (d Device) FilterValue() string {
	return d.Name + " " + d.Description
}

// DetectBootType reads the first sector of a device and determines
// whether it contains a bootable MBR, GPT, or neither.
func DetectBootType(devicePath string) BootType {
	buf, err := readFirstSector(devicePath)
	if err != nil || len(buf) < 512 {
		return BootNone
	}

	// Check MBR boot signature (last 2 bytes = 0x55AA)
	hasMBRSig := buf[510] == 0x55 && buf[511] == 0xAA

	// Check for GPT: "EFI PART" at offset 0 of LBA 1 (byte 512)
	// But we only read 512 bytes, so check the MBR partition table
	// for a GPT protective partition (type 0xEE)
	hasGPTProtective := false
	if hasMBRSig {
		// MBR partition table starts at offset 446, each entry is 16 bytes
		for i := 0; i < 4; i++ {
			partType := buf[446+i*16+4]
			if partType == 0xEE {
				hasGPTProtective = true
				break
			}
		}
	}

	if hasGPTProtective {
		return BootGPT
	}

	if hasMBRSig {
		// Check if any partition entry has a non-zero type (actual partitions exist)
		for i := 0; i < 4; i++ {
			partType := buf[446+i*16+4]
			if partType != 0x00 {
				return BootMBR
			}
		}
		// Has signature but no partitions — likely bootable (e.g., dd'd ISO)
		// Check if the boot code area (bytes 0-445) is non-zero
		nonZero := false
		for _, b := range buf[:446] {
			if b != 0 {
				nonZero = true
				break
			}
		}
		if nonZero {
			return BootMBR
		}
	}

	return BootNone
}
