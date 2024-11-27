package metric_functions

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func GetCPUUsage() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getCPUUsageLinux()
	case "darwin":
		return getCPUUsageMacOS()
	case "windows":
		return getCPUUsageWindows()
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// Linux: Uses top or mpstat
func getCPUUsageLinux() (float64, error) {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)'")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(output))
	if len(fields) > 7 {
		idleStr := strings.TrimSuffix(fields[7], "%")
		idle, err := strconv.ParseFloat(idleStr, 64)
		if err != nil {
			return 0, err
		}
		return 100.0 - idle, nil
	}
	return 0, fmt.Errorf("could not parse top output")
}

// MacOS: Uses iostat
func getCPUUsageMacOS() (float64, error) {
	cmd := exec.Command("iostat", "-c", "1", "2")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, " ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				idle, err := strconv.ParseFloat(fields[len(fields)-1], 64)
				if err != nil {
					return 0, err
				}
				return 100.0 - idle, nil
			}
		}
	}
	return 0, fmt.Errorf("could not parse iostat output")
}

// Windows: Uses wmic
func getCPUUsageWindows() (float64, error) {
	cmd := exec.Command("wmic", "cpu", "get", "loadpercentage")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "LoadPercentage" {
			usage, err := strconv.ParseFloat(line, 64)
			if err != nil {
				return 0, err
			}
			return usage, nil
		}
	}
	return 0, fmt.Errorf("could not parse wmic output")
}

// func main() {
// 	usage, err := getCPUUsage()
// 	if err != nil {
// 		fmt.Println("Error getting CPU usage:", err)
// 	} else {
// 		fmt.Printf("CPU Usage: %.2f%%\n", usage)
// 	}
// }
