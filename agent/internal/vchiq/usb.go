package vchiq

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

type USBDevice struct {
	Bus     string `json:"bus"`
	Device  string `json:"device"`
	ID      string `json:"id"`
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
}

var ErrExecutingLsusb = errors.New("error executing lsusb")

func GetUSBList() ([]USBDevice, error) {
	out, err := exec.Command("lsusb").Output()
	if err != nil {
		return nil, ErrExecutingLsusb
	}
	output := string(out)

	re := regexp.MustCompile(`Bus (\d+) Device (\d+): ID (\w+):(\w+) (.+)`)
	var devices []USBDevice

	for _, line := range strings.Split(output, "\n") {
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			devices = append(devices, USBDevice{
				Bus:     matches[1],
				Device:  matches[2],
				ID:      matches[3] + ":" + matches[4],
				Vendor:  matches[3],
				Product: matches[4] + " " + matches[5],
			})
		}
	}
	return devices, nil
}