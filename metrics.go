package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"uptinio-server-agent/metric_functions"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
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
	// Get the MAC address
	macAddress, err := metric_functions.GetMacAddress()
	if err != nil {
		macAddress = "unknown" // Default if unable to retrieve the MAC address
	}

	// Get the hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get the private IP address
	privateIP := metric_functions.GetPrivateIP()

	// Get the public IP address
	publicIP := metric_functions.GetPublicIP()

	// Get the CPU model information
	cpuInfo, err := cpu.Info()
	cpuModel := "unknown"
	if err == nil && len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	// Get the operating system
	operatingSystem := runtime.GOOS

	// Get the system uptime (seconds)
	uptime, err := host.Uptime()
	if err != nil {
		uptime = 0
	}

	// Get the kernel version
	kernelVersion, err := host.KernelVersion()
	if err != nil {
		kernelVersion = "unknown"
	}

	return map[string]interface{}{
		"public_ip":        publicIP,
		"private_ip":       privateIP,
		"hostname":         hostname,
		"mac_address":      macAddress,
		"cpu_cores":        runtime.NumCPU(),
		"cpu_model":        cpuModel,
		"operating_system": operatingSystem,
		"uptime":           int(uptime),
		"kernel_version":   kernelVersion,
	}
}
