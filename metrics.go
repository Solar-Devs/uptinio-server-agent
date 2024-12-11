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
	"github.com/shirou/gopsutil/net"
)

func collectMetrics() ([]Metric, []error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var metrics []Metric
	var errors []error

	// CPU
	cpuUsage, err := metric_functions.GetCPUUsageAWSUnits()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting CPU usage: %w", err))
	} else {
		metrics = append(metrics, Metric{Metric: "cpu_used", Value: cpuUsage, Timestamp: now})
	}

	// Memory
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting memory stats: %w", err))
	} else {
		memoryUsage := float64(vmStats.Used)
		metrics = append(metrics, Metric{Metric: "mem_used_b", Value: memoryUsage, Timestamp: now})
	}

	// Disk Usage
	diskStats, err := disk.Usage("/")
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting disk stats: %w", err))
	} else {
		diskUsage := float64(diskStats.Used)
		metrics = append(metrics, Metric{Metric: "disk_used_b", Value: diskUsage, Timestamp: now})
	}

	// Network Metrics
	netStats, err := net.IOCounters(false)
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting network stats: %w", err))
	} else if len(netStats) > 0 {
		metrics = append(metrics, Metric{
			Metric:    "net_sent_b",
			Value:     float64(netStats[0].BytesSent), // Total data sent in bytes since uptime
			Timestamp: now,
		})
		metrics = append(metrics, Metric{
			Metric:    "net_recv_b",
			Value:     float64(netStats[0].BytesRecv), // Total data received in bytes since uptime
			Timestamp: now,
		})

		metrics = append(metrics, Metric{
			Metric:    "pkt_sent",
			Value:     float64(netStats[0].PacketsSent), // Sent packets since uptime
			Timestamp: now,
		})
		metrics = append(metrics, Metric{
			Metric:    "pkt_recv",
			Value:     float64(netStats[0].PacketsRecv), // Received packets since uptime
			Timestamp: now,
		})
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

	diskStats, err := disk.Usage("/")
	disk_total_bytes := uint64(0)
	if err == nil {
		disk_total_bytes = diskStats.Total
	}

	vmStats, err := mem.VirtualMemory()
	memory_total_bytes := uint64(0)
	if err == nil {
		memory_total_bytes = vmStats.Total
	}

	return map[string]interface{}{
		"public_ip":          publicIP,
		"private_ip":         privateIP,
		"hostname":           hostname,
		"mac_address":        macAddress,
		"cpu_cores":          runtime.NumCPU(),
		"cpu_model":          cpuModel,
		"operating_system":   operatingSystem,
		"uptime":             int(uptime),
		"kernel_version":     kernelVersion,
		"disk_total_bytes":   disk_total_bytes,
		"memory_total_bytes": memory_total_bytes,
	}
}
