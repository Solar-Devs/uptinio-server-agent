package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

const HOST_PATH = "api/v1/server_metrics"

func buildURL(schema, host, hostPath string) (string, error) {
	u := &url.URL{
		Scheme: schema,
		Host:   host,
	}

	u.Path = path.Join(u.Path, hostPath)

	return u.String(), nil
}

func sendMetrics(payload Payload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}
	fullURL, err := buildURL(config.Schema, config.Host, HOST_PATH)
	if err != nil {
		return fmt.Errorf("Error building URL: %v\n", err)
	}
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
