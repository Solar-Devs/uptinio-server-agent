package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEffectiveCommandPollInterval(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 15, effectiveCommandPollInterval(0))
	assert.Equal(t, 30, effectiveCommandPollInterval(30))
	assert.Equal(t, -1, effectiveCommandPollInterval(-1))
}

func TestValidateArgv(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateArgv([]string{"echo", "hi"}))
	require.Error(t, validateArgv(nil))
	require.Error(t, validateArgv([]string{}))
	require.Error(t, validateArgv([]string{"echo", ""}))
	require.Error(t, validateArgv([]string{"echo", "a\x00b"}))

	long := make([]string, maxCommandArgvLen+1)
	for i := range long {
		long[i] = "x"
	}
	require.Error(t, validateArgv(long))
}

func TestValidateAndBuildEnv(t *testing.T) {
	t.Parallel()

	env, err := validateAndBuildEnv(map[string]string{"UPTINIO_SECRET": "s3cr3t"})
	require.NoError(t, err)
	assert.Contains(t, env, "UPTINIO_SECRET=s3cr3t")

	_, err = validateAndBuildEnv(map[string]string{"": "x"})
	require.Error(t, err)

	_, err = validateAndBuildEnv(map[string]string{"BAD=KEY": "x"})
	require.Error(t, err)

	_, err = validateAndBuildEnv(map[string]string{"K": "a\x00b"})
	require.Error(t, err)
}

func TestCappedBufferTruncates(t *testing.T) {
	t.Parallel()

	var buf cappedBuffer
	buf.limit = 8
	n, err := buf.Write([]byte("abcdefghij"))
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.True(t, buf.truncated)
	assert.Equal(t, "abcdefgh", buf.String())
}

func TestFetchPendingCommands_AcceptsCommandsAndJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/server_commands/pending", r.URL.Path)
		assert.Equal(t, "board-1", r.URL.Query().Get("motherboard_id"))
		assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"commands": []map[string]interface{}{
				{"uuid": "c1", "argv": []string{"true"}},
			},
			"jobs": []map[string]interface{}{
				{"uuid": "j1", "argv": []string{"true"}},
			},
		})
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := commandsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "Bearer secret"}
	commandsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		commandsHTTPClient = origClient
	})

	cmds, err := fetchPendingCommands("board-1")
	require.NoError(t, err)
	require.Len(t, cmds, 2)
	assert.Equal(t, "c1", cmds[0].UUID)
	assert.Equal(t, "j1", cmds[1].UUID)
}

func TestFetchPendingCommands_JobsOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jobs": []map[string]interface{}{
				{"uuid": "j-only", "argv": []string{"echo", "hi"}},
			},
		})
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := commandsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "token"}
	commandsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		commandsHTTPClient = origClient
	})

	cmds, err := fetchPendingCommands("board-1")
	require.NoError(t, err)
	require.Len(t, cmds, 1)
	assert.Equal(t, "j-only", cmds[0].UUID)
}

func TestPostCommandResult(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/server_commands/abc-123/result", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := commandsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "token"}
	commandsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		commandsHTTPClient = origClient
	})

	err := postCommandResult("abc-123", CommandResult{ExitCode: 0, Output: "ok"})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"exit_code":0`)
	assert.Contains(t, string(gotBody), `"output":"ok"`)
}

func TestExecuteRemoteCommand_SuccessWithEnv(t *testing.T) {
	result := executeRemoteCommand(RemoteCommand{
		UUID: "u1",
		Argv: []string{"/bin/sh", "-c", "printf '%s' \"$UPTINIO_TEST_SECRET\""},
		Env:  map[string]string{"UPTINIO_TEST_SECRET": "injected-value"},
	})
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "injected-value", result.Output)
	assert.False(t, result.TimedOut)
	assert.Empty(t, result.Error)
}

func TestExecuteRemoteCommand_InvalidArgv(t *testing.T) {
	result := executeRemoteCommand(RemoteCommand{UUID: "u1", Argv: nil})
	assert.Equal(t, -1, result.ExitCode)
	assert.Contains(t, result.Error, "argv")
}

func TestExecuteRemoteCommand_Timeout(t *testing.T) {
	result := executeRemoteCommand(RemoteCommand{
		UUID:           "u-timeout",
		Argv:           []string{"sleep", "2"},
		TimeoutSeconds: 1,
	})
	assert.True(t, result.TimedOut)
	assert.Equal(t, -1, result.ExitCode)
	assert.Contains(t, result.Error, "timed out")
}

func TestExecuteRemoteCommand_OutputCap(t *testing.T) {
	// Generate more than 256KiB of output.
	result := executeRemoteCommand(RemoteCommand{
		UUID: "u-cap",
		Argv: []string{"dd", "if=/dev/zero", "bs=1024", "count=300", "status=none"},
	})
	assert.Equal(t, 0, result.ExitCode)
	assert.True(t, result.Truncated)
	assert.LessOrEqual(t, len(result.Output), maxCommandOutputBytes)
}

func TestPollAndRunCommands_EndToEnd(t *testing.T) {
	var posted map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/pending"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"commands": []map[string]interface{}{
					{
						"uuid":            "job-1",
						"argv":            []string{"echo", "hello-remote"},
						"timeout_seconds": 5,
					},
				},
			})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/result"):
			body, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(body, &posted))
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	origConfig := config
	origClient := commandsHTTPClient
	config = Config{Schema: "http", Host: strings.TrimPrefix(server.URL, "http://"), AuthToken: "token"}
	commandsHTTPClient = server.Client()
	t.Cleanup(func() {
		config = origConfig
		commandsHTTPClient = origClient
	})

	// Use a short sleep-free path: pollAndRunCommands resolves motherboard id locally.
	pollAndRunCommands()

	require.NotNil(t, posted)
	assert.EqualValues(t, 0, posted["exit_code"])
	assert.Contains(t, posted["output"], "hello-remote")
}

func TestBuildCommandResultURL(t *testing.T) {
	t.Parallel()
	u, err := buildCommandResultURL("https", "app.example.com", "uuid-1")
	require.NoError(t, err)
	assert.Equal(t, "https://app.example.com/api/v1/server_commands/uuid-1/result", u)
}

func TestExecuteRemoteCommand_NonZeroExit(t *testing.T) {
	result := executeRemoteCommand(RemoteCommand{
		UUID: "u-fail",
		Argv: []string{"false"},
	})
	assert.NotEqual(t, 0, result.ExitCode)
	assert.False(t, result.TimedOut)
	assert.Empty(t, result.Error)
}
