package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	// available flags
	createConfig := flag.Bool("create-config", false, "Create a new configuration")
	authToken := flag.String("auth-token", "", "Authorization token for the agent")
	url := flag.String("url", "", "Metrics server URL")
	collectIntervalSec := flag.Int("collect-interval-sec", int(defaultConfig.CollectInterval.Seconds()), "Metrics collection interval in seconds")
	sendIntervalSec := flag.Int("send-interval-sec", int(defaultConfig.SendInterval.Seconds()), "Metrics sending interval in seconds")
	metricsPath := flag.String("metrics-path", defaultConfig.MetricsPath, "Metrics file path")
	flag.StringVar(&ConfigPath, "config-path", DefaultConfigPath, "Config file path, must be a json file")

	flag.Parse()

	if *createConfig {
		if err := createConfiguration(*authToken, *url, *collectIntervalSec, *sendIntervalSec, *metricsPath); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		return
	}

	config = LoadConfig()

	fmt.Printf("Starting agent with the following configuration:\n"+
		"Metrics File Path: %s\n"+
		"Server URL: %s\n"+
		"Auth Token: %s\n"+
		"Collect Interval: %v\n"+
		"Send Interval: %v\n",
		config.MetricsPath, config.URL, config.AuthToken, config.CollectInterval, config.SendInterval)

	collectTicker := time.NewTicker(config.CollectInterval)
	sendTicker := time.NewTicker(config.SendInterval)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	for {
		select {
		case <-collectTicker.C:
			metrics, errors := collectMetrics()
			if len(errors) > 0 {
				fmt.Println("Errors encountered while collecting metrics:")
				for _, err := range errors {
					fmt.Println(err)
				}
			}

			attributes := getAttributes()

			payload := Payload{
				Attributes: attributes,
				Metrics:    metrics,
			}

			if err := saveMetricsToFile(payload); err != nil {
				fmt.Println("Error saving metrics:", err)
			}

		case <-sendTicker.C:
			fmt.Println("Trying to send metrics to server...")
			payload, err := loadMetricsFromFile()
			if err != nil {
				fmt.Println("Error loading metrics from file:", err)
				continue
			}

			if len(payload.Metrics) == 0 {
				fmt.Println("No metrics available to send")
				continue
			}

			if err := sendMetrics(payload); err != nil {
				fmt.Println("Error sending metrics:", err)
			} else {
				fmt.Println("Metrics succesfully sent... cleaning file")
				_ = clearMetricsFile()
			}
		}
	}
}
