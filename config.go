package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	DefaultConfigPath = filepath.Join(getConfBaseDir(), "uptinio-server-agent", "config.json")
	ConfigPath        string
)

// defaultConfig provides default values for the configuration
var defaultConfig = Config{
	MetricsPath: filepath.Join(getMetricsBaseDir(), "uptinio-server-agent", "metrics.json"), // metrics file path
	//URL:             "http://localhost:80/api/v1/server_metrics",                                // Metrics are sent here
	//AuthToken:       "",                                                                         // Authorization token
	CollectInterval: 60 * time.Second,  // Collect metrics interval
	SendInterval:    600 * time.Second, // Send metrics interval
}

var config Config

// LoadConfig loads the configuration from a file
func LoadConfig() Config {
	// Try to open the configuration file
	file, err := os.Open(ConfigPath)
	if os.IsNotExist(err) {
		// File doesn't exist
		panic("Configuration file not found. " +
			"Please, create one before running the program. " +
			"Use the --create-config flag. " +
			"View more information in the README.")
	} else if err != nil {
		panic(fmt.Sprintf("Error opening configuration file: %v", err))
	}
	defer file.Close()

	// Decode the configuration file
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		panic(fmt.Sprintf("Error decoding configuration file: %v", err))
	}

	return config
}

// saveConfigFile saves the configuration to a file
func saveConfigFile(config Config) error {

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	file, err := os.Create(ConfigPath)
	if err != nil {
		return fmt.Errorf("error creating configuration file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(&config); err != nil {
		return fmt.Errorf("error encoding configuration: %w", err)
	}
	fmt.Printf("Configuration saved to %s\n", ConfigPath)
	return nil
}

// Function to get the base directory where the metrics file will be saved based on the operating system
func getConfBaseDir() string {
	switch runtime.GOOS {
	case "windows":
		// On Windows, the directory structure is:
		// C:\Users\<USERNAME>\AppData\Local
		// Example: C:\Users\JohnDoe\AppData\Local
		return os.Getenv("LOCALAPPDATA")
	case "darwin":
		// On macOS, the directory structure is:
		// /Users/<USERNAME>/Library/Application Support
		// Example: /Users/JohnDoe/Library/Application Support
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default: // Linux or other systems
		// On Linux, the directory structure is:
		// /home/<USERNAME>/.local/share
		// Example: /home/johndoe/.local/share
		return filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
}

func createConfiguration(authToken string, url string, collectIntervalSec int, sendIntervalSec int, metricsPath string) error {
	if authToken == "" {
		return fmt.Errorf("parameter 'auth token' is mandatory")
	}
	if url == "" {
		return fmt.Errorf("parameter 'url' is mandatory")
	}

	config := Config{
		MetricsPath:     metricsPath,
		URL:             url,
		AuthToken:       authToken,
		CollectInterval: time.Duration(collectIntervalSec) * time.Second,
		SendInterval:    time.Duration(sendIntervalSec) * time.Second,
	}

	err := saveConfigFile(config)

	return err
}
