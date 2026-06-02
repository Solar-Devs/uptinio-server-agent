package metric_functions

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeAWSCPUUnits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent float64
		numCPU  int
		want    float64
	}{
		{"idle single core", 0, 1, 0},
		{"half of one core", 50, 1, 512},
		{"full single core", 100, 1, 1024},
		{"half of four cores", 50, 4, 2048},
		{"quarter of eight cores", 25, 8, 2048},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeAWSCPUUnits(tt.percent, tt.numCPU)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCPUUsage_LinuxIntegration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux integration test")
	}

	usage, err := GetCPUUsage()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, usage, 0.0)
	assert.LessOrEqual(t, usage, 100.0)
}

func TestGetCPUUsageAWSUnits_LinuxIntegration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux integration test")
	}

	units, err := GetCPUUsageAWSUnits()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, units, 0.0)

	maxUnits := ComputeAWSCPUUnits(100, runtime.NumCPU())
	assert.LessOrEqual(t, units, maxUnits)
}
