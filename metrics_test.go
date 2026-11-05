package main

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAttributes_LinuxIntegration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux integration test")
	}

	attrs := getAttributes()

	require.NotEmpty(t, attrs["hostname"])
	require.NotEmpty(t, attrs["operating_system"])
	assert.Equal(t, "linux", attrs["operating_system"])

	id, ok := attrs["motherboard_id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)
	assert.LessOrEqual(t, len(id), 256, "motherboard_id must be capped to avoid oversized payloads")
}

func TestCollectMetrics_LinuxIntegration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux integration test")
	}

	metrics, errs := collectMetrics()
	for _, err := range errs {
		t.Logf("collect warning: %v", err)
	}
	require.NotEmpty(t, metrics)

	byName := make(map[string]Metric, len(metrics))
	for _, m := range metrics {
		byName[m.Metric] = m
		assert.NotEmpty(t, m.Timestamp)
		assert.GreaterOrEqual(t, m.Value, 0.0, m.Metric)
	}

	required := []string{
		"cpu_used",
		"mem_used_b",
		"mem_avail_b",
		"swap_used_b",
		"disk_used_b",
		"disk_avail_b",
		"load_1",
		"load_5",
		"load_15",
		"net_sent_b",
		"net_recv_b",
		"pkt_sent",
		"pkt_recv",
	}
	for _, name := range required {
		_, ok := byName[name]
		assert.True(t, ok, "missing metric %s", name)
	}
}
