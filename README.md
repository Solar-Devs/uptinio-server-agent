# Server Monitor Agent

The **Server Monitor Agent** is a lightweight Go-based tool designed to periodically collect server metrics and send them to a specified destination. The agent can be configured to collect data at regular intervals and transmit it at a different frequency, ensuring efficient and reliable monitoring.

---

## Features
- Periodically collects server metrics: CPU, memory, swap, disk, load averages, and network I/O.
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

1. **Download the latest binary**: It fetches the latest Linux release binary from the GitHub repository, storing it in `/usr/local/bin/uptinio-agent`.

2. **Create a configuration file**: It generates a configuration file for the agent based on the provided parameters. By default, the configuration is stored in `/etc/uptinio-agent.yaml`.

3. **Create a service**: The script sets up systemd, OpenRC, or SysV (in that order) as `uptinio-agent`. On systemd you can check status with:

```
systemctl status uptinio-agent
```

### Updating the agent

Re-run the same installer on a host that already has the agent. It detects the existing install, stops the service, replaces the binary, refreshes the unit/init script, and restarts the agent.

Unspecified options are kept from the existing config (including `auth_token`), so this is enough after a new release:

```
sudo bash linux_installer.sh
```

Pass flags only when you want to change settings during the update (for example `--host` or a new `--auth-token`).

The parameters of `linux_installer.sh` are the following:

* `auth-token`: The authorization token used for the request. **Required on first install**; optional on update if already present in the config.
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
      "motherboard_id": "ABC123",
      "mac_address": "00-0a-95-9d-67-16",
      "public_ip": "203.0.113.10",
      "private_ip": "10.0.0.5",
      "hostname": "web-1",
      "cpu_cores": 4,
      "cpu_model": "Intel(R) Xeon(R) CPU",
      "operating_system": "linux",
      "uptime": 86400,
      "kernel_version": "6.8.0",
      "disk_total_bytes": 107374182400,
      "memory_total_bytes": 8589934592
    },
    "metrics": [
      { "metric": "cpu_used", "value": 2048, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "mem_used_b", "value": 4294967296, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "mem_avail_b", "value": 2147483648, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "swap_used_b", "value": 0, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "disk_used_b", "value": 53687091200, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "disk_avail_b", "value": 53687091200, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "load_1", "value": 0.45, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "load_5", "value": 0.52, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "load_15", "value": 0.61, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "net_sent_b", "value": 1048576, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "net_recv_b", "value": 2097152, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "pkt_sent", "value": 1000, "timestamp": "2024-11-06T12:00:00Z" },
      { "metric": "pkt_recv", "value": 2000, "timestamp": "2024-11-06T12:00:00Z" }
    ]
  }'
```

### Metrics

Collected every `$COLLECT_INTERVAL_SEC`:

| Metric | Meaning |
|--------|---------|
| `cpu_used` | CPU usage in AWS-style units (1024 = 1 core at 100%) |
| `mem_used_b` | Used memory in bytes |
| `mem_avail_b` | Available memory in bytes |
| `swap_used_b` | Used swap in bytes |
| `disk_used_b` | Used disk space on `/` in bytes |
| `disk_avail_b` | Available disk space on `/` in bytes |
| `load_1` / `load_5` / `load_15` | Load averages (1 / 5 / 15 minutes) |
| `net_sent_b` / `net_recv_b` | Network bytes sent / received since boot |
| `pkt_sent` / `pkt_recv` | Network packets sent / received since boot |

### Attributes

Sent with each payload: `motherboard_id`, `mac_address`, `public_ip`, `private_ip`, `hostname`, `cpu_cores`, `cpu_model`, `operating_system`, `uptime` (seconds), `kernel_version`, `disk_total_bytes`, `memory_total_bytes`.

The `$URL` variable follows the structure, `$URL=$SCHEMA://$HOST/$HOST_PATH`, where `$SCHEMA` and `$HOST` are configurable values that can be modified in the configuration file. The third component, `$HOST_PATH`, is a static value defined directly in the `sender.go` code.

Every time the request response is a `201` (success), the local metrics file is cleared.
