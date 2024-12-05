package main

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
