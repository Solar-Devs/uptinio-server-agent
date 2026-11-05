package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildURL(t *testing.T) {
	t.Parallel()

	url, err := buildURL("https", "app.example.com", "api/v1/server_metrics")
	require.NoError(t, err)
	assert.Equal(t, "https://app.example.com/api/v1/server_metrics", url)
}

func TestSendMetrics_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "cpu_used")
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := metricsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "Bearer secret-token"}
	metricsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		metricsHTTPClient = origClient
	})

	err := sendMetrics(Payload{
		Version:    "test",
		Attributes: map[string]interface{}{"motherboard_id": "board-1"},
		Metrics:    []Metric{{Metric: "cpu_used", Value: 100, Timestamp: "2026-01-01T00:00:00Z"}},
	})
	require.NoError(t, err)
}

func TestSendMetrics_RetriesOn413WithLatestMetrics(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		body, _ := io.ReadAll(r.Body)
		assert.NotContains(t, string(body), `"metric":"old"`)
		assert.Contains(t, string(body), `"metric":"new"`)
		assert.Contains(t, string(body), `"timestamp":"2026-01-01T00:01:00Z"`)
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := metricsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "token"}
	metricsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		metricsHTTPClient = origClient
	})

	metrics := []Metric{
		{Metric: "old", Value: 1, Timestamp: "2026-01-01T00:00:00Z"},
		{Metric: "old", Value: 2, Timestamp: "2026-01-01T00:00:00Z"},
		{Metric: "new", Value: 3, Timestamp: "2026-01-01T00:01:00Z"},
		{Metric: "new", Value: 4, Timestamp: "2026-01-01T00:01:00Z"},
	}

	err := sendMetrics(Payload{
		Version:    "test",
		Attributes: map[string]interface{}{"motherboard_id": "board-1"},
		Metrics:    metrics,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
}

func TestMetricsWithLatestTimestamp(t *testing.T) {
	t.Parallel()

	assert.Empty(t, metricsWithLatestTimestamp(nil))

	input := []Metric{
		{Metric: "a", Timestamp: "2026-01-01T00:00:00Z"},
		{Metric: "b", Timestamp: "2026-01-01T00:01:00Z"},
		{Metric: "c", Timestamp: "2026-01-01T00:01:00Z"},
		{Metric: "d", Timestamp: "2026-01-01T00:00:30Z"},
	}
	got := metricsWithLatestTimestamp(input)
	require.Len(t, got, 2)
	assert.Equal(t, "b", got[0].Metric)
	assert.Equal(t, "c", got[1].Metric)
}

func TestSendMetrics_413StillFailsWhenPayloadSmall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := metricsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "token"}
	metricsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		metricsHTTPClient = origClient
	})

	err := sendMetrics(Payload{
		Version:    "test",
		Attributes: map[string]interface{}{"motherboard_id": "board-1"},
		Metrics:    []Metric{{Metric: "cpu_used", Value: 1, Timestamp: "2026-01-01T00:00:00Z"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "413")
}

func TestSendMetrics_MissingAuthToken(t *testing.T) {
	origConfig := config
	config = Config{Schema: "http", Host: "example.com", AuthToken: "   "}
	t.Cleanup(func() { config = origConfig })

	err := sendMetrics(Payload{Attributes: map[string]interface{}{"motherboard_id": "x"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication token not configured")
}
