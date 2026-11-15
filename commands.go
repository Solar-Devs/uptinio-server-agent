package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
	"uptinio-server-agent/metric_functions"
)

const (
	commandsPendingPath     = "api/v1/server_commands/pending"
	defaultCommandPollSecs  = 15
	defaultCommandTimeout   = 60 * time.Second
	maxCommandOutputBytes   = 256 * 1024 // 256 KiB
	maxCommandArgvLen       = 64
	maxCommandArgBytes      = 4096
	maxCommandEnvEntries    = 64
	maxCommandEnvKeyBytes   = 256
	maxCommandEnvValueBytes = 64 * 1024
)

var commandsHTTPClient = &http.Client{Timeout: 15 * time.Second}

// RemoteCommand is a single pending job from the control plane.
type RemoteCommand struct {
	UUID           string            `json:"uuid"`
	Argv           []string          `json:"argv"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	Env            map[string]string `json:"env"`
}

type pendingCommandsResponse struct {
	Commands []RemoteCommand `json:"commands"`
	Jobs     []RemoteCommand `json:"jobs"`
}

// CommandResult is posted back after local execution.
type CommandResult struct {
	ExitCode  int    `json:"exit_code"`
	Output    string `json:"output"`
	TimedOut  bool   `json:"timed_out"`
	Truncated bool   `json:"truncated"`
	Error     string `json:"error,omitempty"`
}

func effectiveCommandPollInterval(seconds int) int {
	if seconds == 0 {
		return defaultCommandPollSecs
	}
	return seconds
}

func resolveMotherboardID() string {
	motherboardID, err := metric_functions.GetMotherboardID()
	if err != nil {
		log.Printf("Error obtaining motherboard ID for commands: %v", err)
		motherboardID, err = metric_functions.GetFallbackDeviceID()
		if err != nil {
			log.Printf("Error obtaining fallback device ID for commands: %v", err)
			return "unknown"
		}
	}
	return metric_functions.SanitizeDeviceID(motherboardID)
}

func pollAndRunCommands() {
	motherboardID := resolveMotherboardID()
	commands, err := fetchPendingCommands(motherboardID)
	if err != nil {
		log.Printf("Error polling pending commands: %v", err)
		return
	}
	if len(commands) == 0 {
		return
	}
	log.Printf("Received %d pending command(s)", len(commands))
	for _, cmd := range commands {
		result := executeRemoteCommand(cmd)
		if err := postCommandResult(cmd.UUID, result); err != nil {
			log.Printf("Error posting result for command %s: %v", cmd.UUID, err)
		}
	}
}

func fetchPendingCommands(motherboardID string) ([]RemoteCommand, error) {
	authToken := strings.TrimSpace(config.AuthToken)
	if authToken == "" {
		return nil, fmt.Errorf("authentication token not configured")
	}

	baseURL, err := buildURL(config.Schema, config.Host, commandsPendingPath)
	if err != nil {
		return nil, fmt.Errorf("error building pending-commands URL: %w", err)
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing pending-commands URL: %w", err)
	}
	q := u.Query()
	q.Set("motherboard_id", motherboardID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating pending-commands request: %w", err)
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Accept", "application/json")

	resp, err := commandsHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching pending commands: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("error reading pending-commands response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code polling commands: %d", resp.StatusCode)
	}

	var parsed pendingCommandsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("error decoding pending-commands response: %w", err)
	}

	// Accept either "commands" or "jobs" (or both, concatenated).
	out := make([]RemoteCommand, 0, len(parsed.Commands)+len(parsed.Jobs))
	out = append(out, parsed.Commands...)
	out = append(out, parsed.Jobs...)
	return out, nil
}

func postCommandResult(commandUUID string, result CommandResult) error {
	if strings.TrimSpace(commandUUID) == "" {
		return fmt.Errorf("command uuid is empty")
	}
	authToken := strings.TrimSpace(config.AuthToken)
	if authToken == "" {
		return fmt.Errorf("authentication token not configured")
	}

	fullURL, err := buildCommandResultURL(config.Schema, config.Host, commandUUID)
	if err != nil {
		return fmt.Errorf("error building command-result URL: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling command result: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating command-result request: %w", err)
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := commandsHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error posting command result: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code posting command result: %d", resp.StatusCode)
	}
	return nil
}

func buildCommandResultURL(schema, host, commandUUID string) (string, error) {
	u := &url.URL{
		Scheme: schema,
		Host:   host,
	}
	u.Path = path.Join(u.Path, "api/v1/server_commands", commandUUID, "result")
	return u.String(), nil
}

func executeRemoteCommand(cmd RemoteCommand) CommandResult {
	if strings.TrimSpace(cmd.UUID) == "" {
		return CommandResult{ExitCode: -1, Error: "missing command uuid"}
	}
	if err := validateArgv(cmd.Argv); err != nil {
		return CommandResult{ExitCode: -1, Error: err.Error()}
	}
	env, err := validateAndBuildEnv(cmd.Env)
	if err != nil {
		return CommandResult{ExitCode: -1, Error: err.Error()}
	}

	timeout := defaultCommandTimeout
	if cmd.TimeoutSeconds > 0 {
		timeout = time.Duration(cmd.TimeoutSeconds) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd.Argv[0], cmd.Argv[1:]...)
	execCmd.Env = env

	var buf cappedBuffer
	buf.limit = maxCommandOutputBytes
	execCmd.Stdout = &buf
	execCmd.Stderr = &buf

	err = execCmd.Run()
	result := CommandResult{
		Output:    buf.String(),
		Truncated: buf.truncated,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = -1
		result.Error = "command timed out"
		return result
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result
		}
		result.ExitCode = -1
		result.Error = err.Error()
		return result
	}
	result.ExitCode = 0
	return result
}

func validateArgv(argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("argv must not be empty")
	}
	if len(argv) > maxCommandArgvLen {
		return fmt.Errorf("argv exceeds max length %d", maxCommandArgvLen)
	}
	for i, arg := range argv {
		if arg == "" {
			return fmt.Errorf("argv[%d] is empty", i)
		}
		if strings.IndexByte(arg, 0) >= 0 {
			return fmt.Errorf("argv[%d] contains a null byte", i)
		}
		if len(arg) > maxCommandArgBytes {
			return fmt.Errorf("argv[%d] exceeds max size %d bytes", i, maxCommandArgBytes)
		}
	}
	return nil
}

func validateAndBuildEnv(extra map[string]string) ([]string, error) {
	if len(extra) > maxCommandEnvEntries {
		return nil, fmt.Errorf("env exceeds max entries %d", maxCommandEnvEntries)
	}
	base := os.Environ()
	if len(extra) == 0 {
		return base, nil
	}

	// Later entries win for duplicate keys when the OS reads the env slice;
	// append overrides after the process environment.
	out := make([]string, 0, len(base)+len(extra))
	out = append(out, base...)
	for key, value := range extra {
		if key == "" {
			return nil, fmt.Errorf("env key is empty")
		}
		if strings.IndexByte(key, 0) >= 0 || strings.Contains(key, "=") {
			return nil, fmt.Errorf("invalid env key %q", key)
		}
		if len(key) > maxCommandEnvKeyBytes {
			return nil, fmt.Errorf("env key exceeds max size %d bytes", maxCommandEnvKeyBytes)
		}
		if strings.IndexByte(value, 0) >= 0 {
			return nil, fmt.Errorf("env value for %q contains a null byte", key)
		}
		if len(value) > maxCommandEnvValueBytes {
			return nil, fmt.Errorf("env value for %q exceeds max size %d bytes", key, maxCommandEnvValueBytes)
		}
		out = append(out, key+"="+value)
	}
	return out, nil
}

// cappedBuffer writes until limit bytes, then discards further writes.
type cappedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	remaining := c.limit - c.buf.Len()
	if remaining <= 0 {
		c.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = c.buf.Write(p[:remaining])
		c.truncated = true
		return len(p), nil
	}
	return c.buf.Write(p)
}

func (c *cappedBuffer) String() string {
	return c.buf.String()
}
