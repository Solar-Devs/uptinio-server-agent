package metric_functions

import (
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

// Gets CPU usage like AWS: 1024 units = 1 core usage.
func GetCPUUsageAWSUnits() (float64, error) {
	cpuUsagePercent, err := GetCPUUsage()
	if err != nil {
		return 0, err
	}

	return ComputeAWSCPUUnits(cpuUsagePercent, runtime.NumCPU()), nil
}

// ComputeAWSCPUUnits converts usage percent to AWS-style CPU units (1024 = one core at 100%).
func ComputeAWSCPUUnits(usagePercent float64, numCPU int) float64 {
	return math.Round((usagePercent * float64(numCPU) * 1024) / 100)
}

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

// Linux: reads /proc/stat via gopsutil (avoids brittle top output formats).
func getCPUUsageLinux() (float64, error) {
	percentages, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		return 0, err
	}
	if len(percentages) == 0 {
		return 0, fmt.Errorf("no CPU usage data")
	}
	return percentages[0], nil
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
