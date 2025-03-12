package metric_functions

import (
	"fmt"
	"net"
	"runtime"
	"strings"
)

func GetMacAddress() (string, error) {
	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		interfaces, err := net.Interfaces()
		if err != nil {
			return "", fmt.Errorf("could not get network interfaces: %w", err)
		}
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			mac := strings.TrimSpace(iface.HardwareAddr.String())
			if mac != "" {
				return mac, nil
			}
		}
		return "", fmt.Errorf("no valid MAC address found")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
