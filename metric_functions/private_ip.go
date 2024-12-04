package metric_functions

import (
	"net"
)

func GetPrivateIP() string {
	// Iterate over network interfaces to find the private IP address
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				// Check if the address is IPv4 and not a loopback address
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}
	return "unknown"
}
