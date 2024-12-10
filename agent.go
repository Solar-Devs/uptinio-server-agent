package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var Version = "unknown" // fill this value when compiling with a flag: -ldflags "-X main.Version=VERSION_VALUE"

func main() {
	// available flags
	getVersion := flag.Bool("version", false, "Get agent version")
	flag.StringVar(&ConfigPath, "config-path", "", "Config file path, must be a yaml file")

	flag.Parse()

	if *getVersion {
		fmt.Printf("%s\n", Version)
		return
	}

	if ConfigPath == "" {
		fmt.Printf("parameter 'config-path' is mandatory")
		return
	}

	config = LoadConfig()

	fmt.Printf("Starting agent (version: %s) with the following configuration:\n", Version)
	printConfig(config)

	logWriter, err := NewSizeLimitedLogWriter(config.LogPath, config.MaxLogSizeMB, int(float64(config.MaxLogSizeMB)*0.9))
	if err != nil {
		panic(fmt.Sprintf("Error setting up log file: %v", err))
	}
	defer logWriter.Close()
	log.SetOutput(logWriter)

	collectTicker := time.NewTicker(time.Duration(config.CollectIntervalInSeconds) * time.Second)
	sendTicker := time.NewTicker(time.Duration(config.SendIntervalInSeconds) * time.Second)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	for {
		select {
		case <-collectTicker.C:
			metrics, errors := collectMetrics()
			if len(errors) > 0 {
				log.Println("Errors encountered while collecting metrics:")
				for _, err := range errors {
					log.Println(err)
				}
			}

			attributes := getAttributes()

			payload := Payload{
				Version:    Version,
				Attributes: attributes,
				Metrics:    metrics,
			}

			if err := saveMetricsToFile(payload); err != nil {
				log.Println("Error saving metrics:", err)
			}

		case <-sendTicker.C:
			log.Println("Trying to send metrics to server...")
			payload, err := loadMetricsFromFile()
			if err != nil {
				log.Println("Error loading metrics from file:", err)
				continue
			}

			if len(payload.Metrics) == 0 {
				log.Println("No metrics available to send")
				continue
			}

			if err := sendMetrics(payload); err != nil {
				log.Println("Error sending metrics:", err)
			} else {
				log.Println("Metrics succesfully sent... cleaning metrics file")
				_ = clearMetricsFile()
			}
		}
	}
}
