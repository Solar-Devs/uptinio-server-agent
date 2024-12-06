package main

import (
	"os"
	"sync"
)

type Metric struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

type Payload struct {
	Version    string                 `json:"agent_version"`
	Attributes map[string]interface{} `json:"attributes"`
	Metrics    []Metric               `json:"metrics"`
}

// Config holds the application configuration
type Config struct {
	MetricsPath              string `yaml:"metrics_path"`
	LogPath                  string `yaml:"log_path"`
	MaxLogSizeMB             int    `yaml:"max_log_file_size_in_MB"`
	Schema                   string `yaml:"schema"`
	Host                     string `yaml:"host"`
	AuthToken                string `yaml:"auth_token"`
	CollectIntervalInSeconds int    `yaml:"collect_interval_in_seconds"`
	SendIntervalInSeconds    int    `yaml:"send_interval_in_seconds"`
}

// SizeLimitedLogWriter is a custom writer that ensures a log file remains within a specified size limit.
type SizeLimitedLogWriter struct {
	filePath   string     // Path to the log file
	maxSize    int64      // Maximum file size in bytes
	keepBytes  int64      // Number of recent bytes to retain when truncating
	currentLog *os.File   // The current log file
	mu         sync.Mutex // Mutex to ensure thread-safe operations
}
