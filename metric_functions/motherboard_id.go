package metric_functions

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
)

// GetMotherboardID returns the motherboard ID of the server
// It uses dmidecode to get the motherboard ID
// If dmidecode is not installed, it returns an error
func GetMotherboardID() (string, error) {
	switch runtime.GOOS {
	case "linux":
		out, err := exec.Command("sudo", "dmidecode", "-s", "baseboard-serial-number").Output()
		if err == nil {
			id := strings.TrimSpace(string(out))
			if id != "" && id != "Not Specified" && id != "Unknown" {
				return id, nil
			}
		}
		out, err = exec.Command("sudo", "dmidecode", "-s", "system-uuid").Output()
		if err == nil {
			id := strings.TrimSpace(string(out))
			if id != "" && id != "Not Specified" && id != "Unknown" {
				return id, nil
			}
		}
		data, err := ioutil.ReadFile("/sys/class/dmi/id/board_serial")
		if err == nil {
			id := strings.TrimSpace(string(data))
			if id != "" && id != "Not Specified" && id != "Unknown" {
				return id, nil
			}
		}
		data, err = ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
		if err == nil {
			id := strings.TrimSpace(string(data))
			if id != "" {
				return id, nil
			}
		}
		return "", fmt.Errorf("motherboard ID not found after multiple attempts")
	case "windows":
		out, err := exec.Command("wmic", "baseboard", "get", "serialnumber").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && strings.ToLower(line) != "serialnumber" {
					return line, nil
				}
			}
		}
		return "", fmt.Errorf("motherboard ID not found on Windows")
	case "darwin":
		out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Serial Number") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						id := strings.TrimSpace(parts[1])
						if id != "" && id != "Not Available" {
							return id, nil
						}
					}
				}
			}
		}
		return "", fmt.Errorf("motherboard ID not found on macOS")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
