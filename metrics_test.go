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
