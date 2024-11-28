package main

import (
	"fmt"
	"time"
)

func main() {

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
