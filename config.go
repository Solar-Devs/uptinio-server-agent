package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

var (
	DefaultConfigPath = filepath.Join(getConfBaseDir(), "uptinio-server-agent", "config.yaml")
	ConfigPath        string
)

// defaultConfig provides default values for the configuration
var defaultConfig = Config{
	MetricsPath:              filepath.Join(getMetricsBaseDir(), "uptinio-server-agent", "metrics.json"), // metrics file path
	Schema:                   "https",                                                                    // request schema
	Host:                     "api.staging.uptinio.com",                                                  // server host
	CollectIntervalInSeconds: 60,                                                                         // Collect metrics interval in seconds
	SendIntervalInSeconds:    60,                                                                         // Send metrics interval in seconds
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
	decoder := yaml.NewDecoder(file)
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

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(&config); err != nil {
		return fmt.Errorf("error encoding configuration: %w", err)
	}
	fmt.Printf("Configuration saved to %s:\n", ConfigPath)
	printConfig(config)
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
		return "/etc"
	}
}

func createConfiguration(
	authToken string, schema string, host string,
	collectIntervalSec int, sendIntervalSec int, metricsPath string) error {
	if authToken == "" {
		return fmt.Errorf("parameter 'auth token' is mandatory")
	}

	if !isValidSchema(schema) {
		return fmt.Errorf("the schema '%s' is invalid. It must be one of: %v\n", schema, VALID_SCHEMAS)
	}

	if hasSchema(host) {
		return fmt.Errorf("the host '%s' must not include a schema (like 'http://' or 'https://').\n", host)
	}

	config := Config{
		MetricsPath:              metricsPath,
		Schema:                   schema,
		Host:                     host,
		AuthToken:                authToken,
		CollectIntervalInSeconds: collectIntervalSec,
		SendIntervalInSeconds:    sendIntervalSec,
	}

	err := saveConfigFile(config)

	return err
}

// printConfig prints the configuration in a readable YAML format
func printConfig(config Config) error {
	// Marshal the configuration into YAML format
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %w", err)
	}

	// Print the YAML to the standard output
	fmt.Println(string(yamlData))
	return nil
}
