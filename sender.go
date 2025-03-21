package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
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
	if _, ok := payload.Attributes["motherboard_id"]; !ok {
		log.Printf("WARNING: motherboard_id not found in attributes")
	} else {
		log.Printf("DEBUG: motherboard_id found: %v", payload.Attributes["motherboard_id"])
	}

	if _, ok := payload.Attributes["mac_address"]; ok {
		log.Printf("WARNING: mac_address still present in attributes")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}
	fullURL, err := buildURL(config.Schema, config.Host, HOST_PATH)
	if err != nil {
		return fmt.Errorf("error building URL: %v", err)
	}

	authToken := strings.TrimSpace(config.AuthToken)
	if authToken == "" {
		return fmt.Errorf("authentication token not configured")
	}

	log.Printf("DEBUG: Schema=%q, Host=%q", config.Schema, config.Host)
	log.Printf("DEBUG: AuthToken length=%d (first 5 characters: %s...)",
		len(authToken), authToken[:min(5, len(authToken))])
	log.Printf("DEBUG: POST URL: %q", fullURL)
	log.Printf("DEBUG: Payload: %s", string(data))

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", authToken)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
