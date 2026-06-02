//go:build linux

package metric_functions

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMotherboardID_LinuxIntegration(t *testing.T) {
	if _, err := exec.LookPath("dmidecode"); err != nil {
		t.Skip("dmidecode not installed")
	}

	id, err := GetMotherboardID()
	if err != nil {
		t.Skipf("motherboard ID not available in this environment: %v", err)
	}

	require.NotEmpty(t, id)
	require.LessOrEqual(t, len(id), 256)
}
