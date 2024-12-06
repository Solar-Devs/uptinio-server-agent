package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ConfigPath string
	config     Config
)

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
