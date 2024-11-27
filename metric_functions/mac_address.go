package metric_functions

import (
	"fmt"
	"net"
)

// Function to get the MAC address dynamically
func GetMacAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("could not get network interfaces: %w", err)
	}

	// Loop through the interfaces to find the first non-loopback one with a valid MAC address
	for _, iface := range interfaces {
		if iface.HardwareAddr.String() != "" && iface.Flags&net.FlagUp != 0 {
			// Return the first found MAC address
			return iface.HardwareAddr.String(), nil
		}
	}

	// Return an error if no valid MAC address is found
	return "", fmt.Errorf("no MAC address found")
}
