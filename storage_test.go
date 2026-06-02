package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveMetricsToFile_CapsStoredMetrics(t *testing.T) {
	dir := t.TempDir()
	metricsPath := filepath.Join(dir, "metrics.json")

	origConfig := config
	config = Config{MetricsPath: metricsPath}
	t.Cleanup(func() { config = origConfig })

	metric := func(name string) Metric {
		return Metric{Metric: name, Value: 1, Timestamp: "2026-01-01T00:00:00Z"}
	}

	// Fill beyond cap (each save adds one metric).
	for i := 0; i < maxStoredMetrics+10; i++ {
		err := saveMetricsToFile(Payload{
			Version:    "test",
			Attributes: map[string]interface{}{"host": "test"},
			Metrics:    []Metric{metric("cpu_used")},
		})
		require.NoError(t, err)
	}

	loaded, err := loadMetricsFromFile()
	require.NoError(t, err)
	assert.Len(t, loaded.Metrics, maxStoredMetrics)
}

func TestSaveAndLoadMetricsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	metricsPath := filepath.Join(dir, "metrics.json")

	origConfig := config
	config = Config{MetricsPath: metricsPath}
	t.Cleanup(func() { config = origConfig })

	payload := Payload{
		Version: "v1",
		Attributes: map[string]interface{}{
			"motherboard_id": "test-board",
		},
		Metrics: []Metric{
			{Metric: "cpu_used", Value: 512, Timestamp: "2026-01-01T00:00:00Z"},
		},
	}

	require.NoError(t, saveMetricsToFile(payload))

	loaded, err := loadMetricsFromFile()
	require.NoError(t, err)
	assert.Equal(t, payload.Version, loaded.Version)
	assert.Equal(t, payload.Attributes["motherboard_id"], loaded.Attributes["motherboard_id"])
	require.Len(t, loaded.Metrics, 1)
	assert.Equal(t, payload.Metrics[0].Metric, loaded.Metrics[0].Metric)
}

func TestClearMetricsFile(t *testing.T) {
	dir := t.TempDir()
	metricsPath := filepath.Join(dir, "metrics.json")

	origConfig := config
	config = Config{MetricsPath: metricsPath}
	t.Cleanup(func() { config = origConfig })

	require.NoError(t, os.WriteFile(metricsPath, []byte(`{"metrics":[]}`), 0o644))
	require.NoError(t, clearMetricsFile())
	_, err := os.Stat(metricsPath)
	assert.True(t, os.IsNotExist(err))
}

func TestLoadMetricsFromFile_MissingFile(t *testing.T) {
	dir := t.TempDir()
	metricsPath := filepath.Join(dir, "does-not-exist.json")

	origConfig := config
	config = Config{MetricsPath: metricsPath}
	t.Cleanup(func() { config = origConfig })

	payload, err := loadMetricsFromFile()
	require.NoError(t, err)
	assert.Empty(t, payload.Metrics)

	raw, err := os.ReadFile(metricsPath)
	require.Error(t, err)
	assert.Nil(t, raw)
}

func TestMetricsFileJSONShape(t *testing.T) {
	dir := t.TempDir()
	metricsPath := filepath.Join(dir, "metrics.json")

	origConfig := config
	config = Config{MetricsPath: metricsPath}
	t.Cleanup(func() { config = origConfig })

	require.NoError(t, saveMetricsToFile(Payload{
		Version:    "v1",
		Attributes: map[string]interface{}{"motherboard_id": "abc"},
		Metrics:    []Metric{{Metric: "mem_used_b", Value: 1024, Timestamp: "2026-01-01T00:00:00Z"}},
	}))

	raw, err := os.ReadFile(metricsPath)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Contains(t, decoded, "metrics")
	assert.Contains(t, decoded, "attributes")
}
