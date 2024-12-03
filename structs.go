package main

type Metric struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

type Payload struct {
	Attributes map[string]interface{} `json:"attributes"`
	Metrics    []Metric               `json:"metrics"`
}

// Config holds the application configuration
type Config struct {
	MetricsPath              string `yaml:"metrics_path"`
	Schema                   string `yaml:"schema"`
	Host                     string `yaml:"host"`
	AuthToken                string `yaml:"auth_token"`
	CollectIntervalInSeconds int    `yaml:"collect_interval_in_seconds"`
	SendIntervalInSeconds    int    `yaml:"send_interval_in_seconds"`
}
