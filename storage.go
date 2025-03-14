package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Save metrics to file
func saveMetricsToFile(newPayload Payload) error {
	filePath := config.MetricsPath

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	existingPayload, _ := loadMetricsFromFile()

	// Combine existing metrics with the new ones
	existingPayload.Metrics = append(existingPayload.Metrics, newPayload.Metrics...)
	existingPayload.Attributes = newPayload.Attributes
	existingPayload.Version = newPayload.Version

	// Create or overwrite the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(existingPayload); err != nil {
		return fmt.Errorf("error encoding metrics: %w", err)
	}
	log.Println("Saved metrics to file at:", filePath)
	return nil
}

// Load metrics from file
func loadMetricsFromFile() (Payload, error) {
	filePath := config.MetricsPath // Get the full file path

	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return Payload{}, nil // File does not exist, return an empty payload
	} else if err != nil {
		return Payload{}, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var payload Payload
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&payload); err != nil {
		return Payload{}, fmt.Errorf("error decoding file: %w", err)
	}
	return payload, nil
}

// Clear the file by deleting it
func clearMetricsFile() error {
	filePath := config.MetricsPath // Get the full file path
	return os.Remove(filePath)
}
