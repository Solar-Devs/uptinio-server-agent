package main

import (
	"fmt"
	"time"
)

const (
	serverURL       = "http://localhost:80/api/v1/server_metrics"
	token           = "0c873d920231f78befedd9d7c5a8f8b2"
	collectInterval = 60 * time.Second
	sendInterval    = 600 * time.Second
)

func main() {
	collectTicker := time.NewTicker(collectInterval) // Cada minuto para recopilar métricas.
	sendTicker := time.NewTicker(sendInterval)       // Cada 10 minutos para enviar métricas.
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

			if err := sendMetrics(serverURL, token, payload); err != nil {
				fmt.Println("Error sending metrics:", err)
			} else {
				fmt.Println("Metrics succesfully sent... cleaning file")
				_ = clearFile()
			}
		}
	}
}
