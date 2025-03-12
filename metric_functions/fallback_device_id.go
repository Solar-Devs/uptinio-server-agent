// This file is used to get the device ID of the server
// It is used as a fallback if the motherboard ID is not found (for example in a VM)
// It uses the machine-id, product_uuid or the combination of hostname and cpuinfo as motherboard_id
package metric_functions

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

func GetFallbackDeviceID() (string, error) {
	data, err := ioutil.ReadFile("/etc/machine-id")
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data)), nil
	}
	
	data, err = ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data)), nil
	}
	
	hostname, _ := exec.Command("hostname").Output()
	cpuinfo, _ := ioutil.ReadFile("/proc/cpuinfo")
	
	if len(hostname) > 0 && len(cpuinfo) > 0 {
		cpuString := string(cpuinfo)
		lines := strings.Split(cpuString, "\n")
		cpuID := ""
		
		for _, line := range lines {
			if strings.Contains(line, "serial") || strings.Contains(line, "processor") {
				cpuID += line
			}
		}
		
		return fmt.Sprintf("%s-%s", strings.TrimSpace(string(hostname)), 
						  strings.ReplaceAll(cpuID, " ", "")), nil
	}
	
	return "", fmt.Errorf("failed to get fallback device ID")
} 