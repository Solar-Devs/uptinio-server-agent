package main

import (
	"fmt"
	"runtime"
	"time"
	"uptinio-server-agent/metric_functions"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

func collectMetrics() ([]Metric, []error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var metrics []Metric
	var errors []error

	// CPU
	cpuUsage, err := metric_functions.GetCPUUsage()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting CPU usage: %w", err))
	} else {
		metrics = append(metrics, Metric{Metric: "cpu", Value: cpuUsage, Timestamp: now})
	}

	// Memory
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting memory stats: %w", err))
	} else {
		memoryUsage := vmStats.UsedPercent
		metrics = append(metrics, Metric{Metric: "memory", Value: memoryUsage, Timestamp: now})
	}

	// Disk
	diskStats, err := disk.Usage("/")
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting disk stats: %w", err))
	} else {
		diskUsage := diskStats.UsedPercent
		metrics = append(metrics, Metric{Metric: "disk", Value: diskUsage, Timestamp: now})
	}

	// Return metrics and any errors encountered
	return metrics, errors
}

func getAttributes() map[string]interface{} {
	macAddress, err := metric_functions.GetMacAddress()
	if err != nil {
		macAddress = "unknown" // Default if unable to get MAC address
	}

	return map[string]interface{}{
		"mac_address": macAddress,       // En el futuro, podrías hacerlo dinámico.
		"cpu_cores":   runtime.NumCPU(), // Núcleos detectados.
	}
}

// func main() {
// 	fmt.Println(collectMetrics())
// 	fmt.Println(getAttributes())
// }
