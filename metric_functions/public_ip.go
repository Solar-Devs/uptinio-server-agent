package metric_functions

import (
	"os/exec"
)

func GetPublicIP() string {
	// Use the curl command to fetch the public IP address
	out, err := exec.Command("curl", "-s", "https://api.ipify.org").Output()
	if err != nil {
		return "unknown"
	}
	return string(out)
}
