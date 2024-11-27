package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const filePath = "metrics.json"

func saveMetricsToFile(newPayload Payload) error {
	existingPayload, _ := loadMetricsFromFile()

	// Combine existing metrics with new ones
	existingPayload.Metrics = append(existingPayload.Metrics, newPayload.Metrics...)
	existingPayload.Attributes = newPayload.Attributes

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(existingPayload); err != nil {
		return fmt.Errorf("error decoding metrics: %w", err)
	}
	fmt.Println("Saved metrics to file")
	return nil
}

func loadMetricsFromFile() (Payload, error) {
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return Payload{}, nil // File doest not exist, empty payload
	} else if err != nil {
		return Payload{}, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	var payload Payload
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&payload); err != nil {
		return Payload{}, fmt.Errorf("error leyendo archivo: %w", err)
	}
	return payload, nil
}

func clearFile() error {
	return os.Remove(filePath)
}
