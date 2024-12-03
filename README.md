# Server Monitor Agent

The **Server Monitor Agent** is a lightweight Go-based tool designed to periodically collect server metrics and send them to a specified destination. The agent can be configured to collect data at regular intervals and transmit it at a different frequency, ensuring efficient and reliable monitoring.

---

## Features
- Periodically collects server metrics such as CPU usage, memory consumption, and more.
- Transmits the collected metrics to a specified destination server.
- Saves metrics locally in case of transmission failure for future retries.
- Simple and efficient implementation in Go.

---

## Generate binary

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

## Generate (overwrite) a config file

The config file is used by the agent to modify certain behaviours during execution.

To generate a config file, you must add the flag `--create-config` when executing the binary. This config file is required to execute the binary normally.


### Example Command
```
./agent --create-config \
  --auth-token "$AUTH_TOKEN" \
  --schema "$SCHEMA" \
  --host "HOST" \
  --collect-interval-in-sec "$COLLECT_INTERVAL_SEC" \
  --send-interval-in-sec "$SEND_INTERVAL_SEC" \
  --metrics-path "$METRICS_PATH" \
  --config-path "$CONFIG_PATH"
```

The variables in the command have the following meanings:

* `auth-token`: The authorization token used for the request. **(Required)**
* `schema`: The schema of the `host`. Default is `https`
* `host`: The host where the collected data will be sent. Default is `api.staging.uptinio.com`
* `collect-interval-sec`: The collection interval in seconds. Default is `60 seconds (1 minute)`
* `send-interval-sec`: The send interval in seconds. Default is `60 seconds (1 minute)`
* `metrics-path`: The path where json metrics are stored before being sent. The default directory depends on the operating system, see `MetricsPath` inside `config.go`. Example value: `/home/johndoe/.local/share/metrics.json`
* `config-path`: The path where the yaml configuration file is generated. The default directory depends on the operating system, see `DefaultConfigPath` inside `config.go`. Example value: `/home/johndoe/.local/share/config.yaml`

Depending on the value of `$CONFIG_PATH`, you might need to run the command with elevated privileges (`sudo`), particularly if the config file needs to be written to a protected directory like `/etc/`.

### Example with `sudo`
```
sudo ./agent --create-config \
  --auth-token "$AUTH_TOKEN" \
  --schema "$SCHEMA" \
  --host "HOST" \
  --collect-interval-in-sec "$COLLECT_INTERVAL_SEC" \
  --send-interval-in-sec "$SEND_INTERVAL_SEC" \
  --metrics-path "$METRICS_PATH" \
  --config-path "$CONFIG_PATH"
```

## Execute agent

Before executing the agent, generate a config file (previous section).
Then, be sure to have execute permissions on binary:

```
chmod +x agent
```

Finally you can run the agent:

```
./agent
```

If you specified a `config-path` when generating a config file, then you should use the same value when executing the agent:

```
./agent --config-path=$CONFIG_PATH
```

## How the data is sent to `$URL`?

The request to `$URL` is made by `sender.go`. It sends the server attributes and metrics stored in `$METRICS_PATH` every `$SEND_INTERVAL_SEC`, and has this structure:

```
  curl -X POST $URL \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
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

# Installing agent with script


# Uninstalling the agent with script
