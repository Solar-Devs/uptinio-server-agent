package main

import "time"

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
	MetricsPath     string        `json:"metrics_path"`
	URL             string        `json:"url"`
	AuthToken       string        `json:"auth_token"`
	CollectInterval time.Duration `json:"collect_interval"`
	SendInterval    time.Duration `json:"send_interval"`
}
