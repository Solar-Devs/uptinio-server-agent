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
	createConfig := flag.Bool("create-config", false, "Create a new configuration")
	getVersion := flag.Bool("get-version", false, "Get agent version")
	getDefaultConfigPath := flag.Bool("get-default-config-path", false, "Get default config path")
	authToken := flag.String("auth-token", "", "Authorization token for the agent")
	schema := flag.String("schema", defaultConfig.Schema, "Schema like http, https...")
	host := flag.String("host", defaultConfig.Host, "host")
	collectIntervalSec := flag.Int("collect-interval-in-sec", int(defaultConfig.CollectIntervalInSeconds), "Metrics collection interval in seconds")
	sendIntervalSec := flag.Int("send-interval-in-sec", int(defaultConfig.SendIntervalInSeconds), "Metrics sending interval in seconds")
	metricsPath := flag.String("metrics-path", defaultConfig.MetricsPath, "Metrics file path")
	logPath := flag.String("log-path", defaultConfig.LogPath, "Log file path")
	flag.StringVar(&ConfigPath, "config-path", DefaultConfigPath, "Config file path, must be a json file")

	flag.Parse()

	if *createConfig {
		if err := createConfiguration(*authToken, *schema, *host, *collectIntervalSec, *sendIntervalSec, *metricsPath, *logPath); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		return
	}

	if *getVersion {
		fmt.Printf("%s\n", Version)
		return
	}

	if *getDefaultConfigPath {
		fmt.Printf("%s\n", DefaultConfigPath)
		return
	}

	config = LoadConfig()

	fmt.Printf("Starting agent (version: %s) with the following configuration:\n", Version)
	printConfig(config)

	logFile, err := getLogFile(*logPath)
	if err != nil {
		panic(fmt.Sprintf("Error setting up log file: %v", err))
	}
	defer logFile.Close()

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
