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
	ConfigPath = filepath.Join(getConfBaseDir(), "uptinio-server-agent-2", "config.json")
)

// defaultConfig provides default values for the configuration
var defaultConfig = Config{
	MetricsPath:     filepath.Join(getMetricsBaseDir(), "uptinio-server-agent", "metrics.json"), // metrics file path
	URL:             "http://localhost:80/api/v1/server_metrics",                                // Metrics are sent here
	AuthToken:       "0c873d920231f78befedd9d7c5a8f8b2",                                         // Authorization token
	CollectInterval: 5 * time.Second,                                                            // Collect metrics interval
	SendInterval:    15 * time.Second,                                                           // Send metrics interval
}

var config Config

// LoadConfig loads the configuration from a file or creates a default one if it doesn't exist
func LoadConfig() Config {
	// Try to open the configuration file
	file, err := os.Open(ConfigPath)
	if os.IsNotExist(err) {
		// File doesn't exist, create a default configuration file
		fmt.Println("Configuration file not found. Creating default configuration.")
		if err := saveConfig(defaultConfig); err != nil {
			fmt.Printf("Error saving default configuration: %v\n", err)
		}
		return defaultConfig
	} else if err != nil {
		fmt.Printf("Error opening configuration file, using default configuration: %v\n", err)
		return defaultConfig
	}
	defer file.Close()

	// Decode the configuration file
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		fmt.Printf("Error decoding configuration file, using default configuration: %v\n", err)
		return defaultConfig
	}

	return config
}

// saveConfig saves the configuration to a file
func saveConfig(config Config) error {

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
