//go:build linux

package usb

import (
	"encoding/json"
	"os/exec"
)

type LsblkOutput struct {
	Blockdevices []struct {
		Name string `json:"name"`
		Rm   bool   `json:"rm"`
		Size uint64 `json:"size"`
		Type string `json:"type"`
	} `json:"blockdevices"`
}

func GetRemovableDevices() ([]Device, error) {
	cmd := exec.Command("lsblk", "-b", "-J", "-o", "NAME,RM,SIZE,TYPE")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var parsed LsblkOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, err
	}

	var devices []Device
	for _, dev := range parsed.Blockdevices {
		if dev.Rm && dev.Type == "disk" {
			devPath := "/dev/" + dev.Name
			devices = append(devices, Device{
				Name:        dev.Name,
				Description: "Removable Drive",
				Size:        dev.Size,
				Path:        devPath,
				Bootable:    DetectBootType(devPath),
			})
		}
	}
	return devices, nil
}
