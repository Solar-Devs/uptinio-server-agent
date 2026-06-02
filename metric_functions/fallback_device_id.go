// This file is used to get the device ID of the server
// It is used as a fallback if the motherboard ID is not found (for example in a VM)
// It uses the machine-id, product_uuid or hostname as motherboard_id
package metric_functions

import (
	"fmt"
	"os"
	"strings"
)

const maxDeviceIDLen = 256

// Overridable in tests.
var (
	fallbackMachineIDPath   = "/etc/machine-id"
	fallbackProductUUIDPath = "/sys/class/dmi/id/product_uuid"
)

func GetFallbackDeviceID() (string, error) {
	if data, err := os.ReadFile(fallbackMachineIDPath); err == nil && len(data) > 0 {
		return SanitizeDeviceID(strings.TrimSpace(string(data))), nil
	}

	if data, err := os.ReadFile(fallbackProductUUIDPath); err == nil && len(data) > 0 {
		return SanitizeDeviceID(strings.TrimSpace(string(data))), nil
	}

	if id, err := fallbackHostnameID(); err == nil {
		return id, nil
	}

	return "", fmt.Errorf("failed to get fallback device ID")
}

var hostnameForFallback = os.Hostname

func fallbackHostnameID() (string, error) {
	hostname, err := hostnameForFallback()
	if err != nil || hostname == "" {
		return "", fmt.Errorf("hostname unavailable")
	}
	return SanitizeDeviceID(hostname), nil
}

func SanitizeDeviceID(id string) string {
	if len(id) <= maxDeviceIDLen {
		return id
	}
	return id[:maxDeviceIDLen]
}
