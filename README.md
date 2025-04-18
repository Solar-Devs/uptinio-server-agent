# Server Monitor Agent

The **Server Monitor Agent** is a lightweight Go-based tool designed to periodically collect server metrics and send them to a specified destination. The agent can be configured to collect data at regular intervals and transmit it at a different frequency, ensuring efficient and reliable monitoring.

---

## Features
- Periodically collects server metrics such as CPU usage, memory consumption, and more.
- Transmits the collected metrics to a specified destination server.
- Saves metrics locally in case of transmission failure for future retries.
- Simple and efficient implementation in Go.

---

## Installing agent with script

### Linux

To install the agent, use the `linux_install.sh` script. The following example demonstrates how to create an agent that sends metrics to `localhost`:

```
sudo bash agent_setup.sh --auth-token $AUTH_TOKEN --host localhost --schema http
```

This script performs the following steps:

1. **Download the latest binary**: It fetches the latest Linux release binary from the GitHub repository, storing it in the `$AGENT_BINARY` directory, which is `/usr/local/bin/uptinio-agent`.

2. **Create a configuration file**: It generates a configuration file for the agent based on the provided parameters. By default, the configutation is stored in `/etc/uptinio-agent.yaml`

3. **Create a systemd service**: The script sets up a systemd service named `uptinio-agent`, which allows you to manage the agent (e.g., check its status) using commands like:

```
systemctl status uptinio-agent
```

The parameters of `agent_setup.sh` are the following:

* `auth-token`: The authorization token used for the request. **(Required)**
* `schema`: The protocol of the `host`. Default is `https`.
* `host`: The host where the collected data will be sent. Default is `app.uptinio.com`
* `collect-interval-sec`: The collection interval in seconds. Default is `60 seconds (1 minute)`
* `send-interval-sec`: The send interval in seconds. Default is `60 seconds (1 minute)`
* `metrics-path`: The path where json metrics are stored before being sent. The default directory is `/var/tmp/uptinio-agent/metrics.json`.
* `log-path`: The path where logs will be stored. The default directory is `/var/log/uptinio-agent/agent.log`.
* `max-log-size-mb`: The maximum size of the log file in megabytes. It will retain only the most recent logs. Default is set to `1024` megabytes (`1` GB).
* `config-path`: The path where the yaml configuration file is generated, by default is `/etc/uptinio-agent.yaml`.
* `uninstall`: This will uninstall the agent.


### Uninstalling the agent

To uninstall the agent, use the `linux_installer.sh` script with the `--uninstall` flag:

```
sudo bash agent_setup.sh --uninstall
```

This script performs the following steps:

1. **Removes the uptinio-agent systemd service**: It stops, disables an removes the systemd service associated with the agent.

2. **Deletes agent files**: The script removes the configuration, metrics and log file of the agent.

3. **Deletes the binary**: The script removes the agent binary from the system.

## Generate agent binary

Clone the repository, move to folder repo and generate your binary depending on your operating system running the corresponding instruction:

Linux
```
GOOS=linux GOARCH=amd64 go build -o agent .
```

Windows
```
GOOS=windows GOARCH=amd64 go build -o agent.exe .
```

MacOS
```
GOOS=darwin GOARCH=amd64 go build -o agent .
```

This generates the binary in the same folder with name `agent`.

## Versioning (optional)
If you want to manage versions, you can store the `$VERSION` value inside `Version` variable during compilation. `Version` default is `unknown`.

Linux
```
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o agent .
```

Windows
```
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o agent.exe .
```

MacOS
```
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o agent .
```

To check your agent's version, run the binary with the `--get-version` flag. This will display the agent's version:

```
./agent --get-version
```

## Execute agent manually

Before executing the agent, generate a yaml config file in your `$CONFIG_PATH`, for example, `/etc/my-config.yaml`, and fill it with your custom values:

```
metrics_path: "$METRICS_PATH"
log_path: "$LOG_PATH"
max_log_file_size_in_MB: $MAX_LOG_SIZE
schema: "$SCHEMA"
host: "$HOST"
auth_token: "$AUTH_TOKEN"
collect_interval_in_seconds: $COLLECT_INTERVAL
send_interval_in_seconds: $SEND_INTERVAL
```

Then, be sure to have execute permissions on binary:

```
chmod +x agent
```

Finally you can run the agent always providing the `$CONFIG_PATH`:

```
./agent --config-path=$CONFIG_PATH
```

You may need to run the agent with `sudo` if `$METRICS_PATH` or `$LOG_PATH` are in protected directories, as write operations and file creation are required.

```
sudo ./agent --config-path=$CONFIG_PATH
```

## How the data is sent to `$URL`?

The request to `$URL` is made by `sender.go`. It sends the agent version, server attributes and metrics stored in `$METRICS_PATH` every `$SEND_INTERVAL_SEC`, and has this structure:

```
  curl -X POST $URL \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_version": "$VERSION",
    "attributes": {
      "mac_address": "00-0a-95-9d-67-16",
      "cpu_cores": 4
      ...
    },
    "metrics": [
     { "metric": "cpu", "value": 70, "timestamp": "2024-11-06T12:00:00Z" },
     { "metric": "memory", "value": 60, "timestamp": "2024-11-06T12:05:00Z" },
     { "metric": "disk", "value": 80, "timestamp": "2024-11-06T12:10:00Z" }
     ...
    ]
  }'
```

Metrics content is collected every `$COLLECT_INTERVAL_SEC`.

The `$URL` variable follows the structure, `$URL=$SCHEMA://$HOST/$HOST_PATH`, where `$SCHEMA` and `$HOST` are configurable values that can be modified in the configuration file. The third component, `$HOST_PATH`, is a static value defined directly in the `sender.go` code.

Every time the request response is a `201` (success), the local metrics file is cleared.
