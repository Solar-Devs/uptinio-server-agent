package metric_functions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeDeviceID(t *testing.T) {
	t.Parallel()

	short := "abc-123"
	assert.Equal(t, short, SanitizeDeviceID(short))

	long := strings.Repeat("x", maxDeviceIDLen+50)
	sanitized := SanitizeDeviceID(long)
	assert.Len(t, sanitized, maxDeviceIDLen)
	assert.Equal(t, long[:maxDeviceIDLen], sanitized)
}

func TestGetFallbackDeviceID_PrefersMachineID(t *testing.T) {
	dir := t.TempDir()
	machineID := filepath.Join(dir, "machine-id")
	require.NoError(t, os.WriteFile(machineID, []byte("debian-test-id\n"), 0o644))

	origMachine, origUUID := fallbackMachineIDPath, fallbackProductUUIDPath
	fallbackMachineIDPath = machineID
	fallbackProductUUIDPath = filepath.Join(dir, "missing-uuid")
	t.Cleanup(func() {
		fallbackMachineIDPath = origMachine
		fallbackProductUUIDPath = origUUID
	})

	id, err := GetFallbackDeviceID()
	require.NoError(t, err)
	assert.Equal(t, "debian-test-id", id)
}

func TestGetFallbackDeviceID_FallsBackToProductUUID(t *testing.T) {
	dir := t.TempDir()
	uuidPath := filepath.Join(dir, "product_uuid")
	require.NoError(t, os.WriteFile(uuidPath, []byte("550e8400-e29b-41d4-a716-446655440000\n"), 0o644))

	origMachine, origUUID := fallbackMachineIDPath, fallbackProductUUIDPath
	fallbackMachineIDPath = filepath.Join(dir, "missing-machine-id")
	fallbackProductUUIDPath = uuidPath
	t.Cleanup(func() {
		fallbackMachineIDPath = origMachine
		fallbackProductUUIDPath = origUUID
	})

	id, err := GetFallbackDeviceID()
	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id)
}

func TestGetFallbackDeviceID_TruncatesLongMachineID(t *testing.T) {
	dir := t.TempDir()
	machineID := filepath.Join(dir, "machine-id")
	longID := strings.Repeat("a", maxDeviceIDLen+100)
	require.NoError(t, os.WriteFile(machineID, []byte(longID), 0o644))

	origMachine := fallbackMachineIDPath
	fallbackMachineIDPath = machineID
	t.Cleanup(func() { fallbackMachineIDPath = origMachine })

	id, err := GetFallbackDeviceID()
	require.NoError(t, err)
	assert.Len(t, id, maxDeviceIDLen)
}

func TestGetFallbackDeviceID_NoSources(t *testing.T) {
	dir := t.TempDir()
	origMachine, origUUID := fallbackMachineIDPath, fallbackProductUUIDPath
	origHostname := hostnameForFallback
	fallbackMachineIDPath = filepath.Join(dir, "missing-machine-id")
	fallbackProductUUIDPath = filepath.Join(dir, "missing-uuid")
	hostnameForFallback = func() (string, error) { return "", fmt.Errorf("unavailable") }
	t.Cleanup(func() {
		fallbackMachineIDPath = origMachine
		fallbackProductUUIDPath = origUUID
		hostnameForFallback = origHostname
	})

	_, err := GetFallbackDeviceID()
	require.Error(t, err)
}

func TestGetFallbackDeviceID_UsesHostname(t *testing.T) {
	dir := t.TempDir()
	origMachine, origUUID := fallbackMachineIDPath, fallbackProductUUIDPath
	origHostname := hostnameForFallback
	fallbackMachineIDPath = filepath.Join(dir, "missing-machine-id")
	fallbackProductUUIDPath = filepath.Join(dir, "missing-uuid")
	hostnameForFallback = func() (string, error) { return "my-server", nil }
	t.Cleanup(func() {
		fallbackMachineIDPath = origMachine
		fallbackProductUUIDPath = origUUID
		hostnameForFallback = origHostname
	})

	id, err := GetFallbackDeviceID()
	require.NoError(t, err)
	assert.Equal(t, "my-server", id)
}
