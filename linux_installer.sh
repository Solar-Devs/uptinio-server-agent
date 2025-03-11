#!/bin/bash

# Check if the script is run as root
if [[ $EUID -ne 0 ]]; then
  echo "This script must be run with sudo or as root." >&2
  exit 1
fi

# Required values
AUTH_TOKEN="" # validation token when sending collected data
# Optional values
HOST="localhost" # URL to send collected data
SCHEMA=http
COLLECT_INTERVAL=60
SEND_INTERVAL=60
METRICS_PATH=/var/tmp/uptinio-agent/metrics.json
LOG_PATH=/var/log/uptinio-agent/agent.log
MAX_LOG_SIZE=1024
CONFIG_PATH=/etc/uptinio-agent.yaml
UNINSTALL=false
# Constants
# BINARY_URL=https://github.com/Solar-Devs/uptinio-server-agent/releases/latest/download/agent-linux-amd64
AGENT_BINARY=/usr/local/bin/uptinio-agent
SERVICE_NAME=uptinio-agent.service
SERVICE_FILE=/etc/systemd/system/$SERVICE_NAME

# Parse arguments
if [[ "$#" -eq 1 && "$1" != "--uninstall" ]]; then # only one argument... the its just auth token
    AUTH_TOKEN="$1"
else # multiple arguments you have to specify --argument value for each
    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --auth-token) AUTH_TOKEN="$2"; shift ;;
            --host) HOST="$2"; shift ;;
            --schema) SCHEMA="$2"; shift ;;
            --collect-interval-in-sec) COLLECT_INTERVAL="$2"; shift ;;
            --send-interval-in-sec) SEND_INTERVAL="$2"; shift ;;
            --metrics-path) METRICS_PATH="$2"; shift ;;
            --log-path) LOG_PATH="$2"; shift ;;
            --max-log-size-mb) MAX_LOG_SIZE="$2"; shift ;;
            --config-path) CONFIG_PATH="$2"; shift ;;
            --uninstall) UNINSTALL=true ;;
            *) echo "Unknown parameter passed: $1"; exit 1 ;;
        esac
        shift
    done
fi


if [ "$UNINSTALL" == "true" ]; then
    echo "Uninstalling uptinio-agent..."

    # Stop and disable the systemd service
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        echo "Stopping systemd service: $SERVICE_NAME"
        systemctl stop "$SERVICE_NAME"
    fi

    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        echo "Disabling systemd service: $SERVICE_NAME"
        systemctl disable "$SERVICE_NAME"
    fi

    # Remove the service file
    if [ -f "$SERVICE_FILE" ]; then
        echo "Removing systemd service file: $SERVICE_FILE"
        rm -f "$SERVICE_FILE"
    fi

    # Reload systemd to apply changes
    systemctl daemon-reload

    # Remove configuration file
    echo "Removing configuration file: $CONFIG_PATH"
    rm "$CONFIG_PATH"

    # Remove metrics file
    echo "Removing metrics file: $METRICS_PATH"
    rm "$METRICS_PATH"

    # Remove logs file
    echo "Removing log file: $LOG_PATH"
    rm "$LOG_PATH"

    # Remove the binary
    if [ -f "$AGENT_BINARY" ]; then
        echo "Removing binary: $AGENT_BINARY"
        rm -f "$AGENT_BINARY"
    fi

    echo "Uninstallation complete."
    exit 0
fi

# Check required parameters
if [ -z "$AUTH_TOKEN" ]; then
    echo "Error: --auth-token is required."
    exit 1
fi

# Install agent binary
echo "Installing agent binary..."
cp ./agent-linux-amd64 "$AGENT_BINARY"
chmod +x "$AGENT_BINARY"
echo "Binary installed at $AGENT_BINARY"

# Run the binary to create the configuration file
echo "Creating configuration file: $CONFIG_PATH..."
cat <<EOF > "$CONFIG_PATH"
metrics_path: "$METRICS_PATH"
log_path: "$LOG_PATH"
max_log_file_size_in_MB: $MAX_LOG_SIZE
schema: "$SCHEMA"
host: "$HOST"
auth_token: "$AUTH_TOKEN"
collect_interval_in_seconds: $COLLECT_INTERVAL
send_interval_in_seconds: $SEND_INTERVAL
EOF

# Create systemd service file
echo "Creating systemd service file..."
cat <<EOF >"$SERVICE_FILE"
[Unit]
Description=Uptinio Server Monitoring Agent
After=network.target

[Service]
ExecStart=$AGENT_BINARY --config-path $CONFIG_PATH
Restart=always
User=$(whoami)

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable the service
echo "Reloading systemd daemon..."
systemctl daemon-reload
echo "Enabling and starting systemd service..."
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

echo "Installing dmidecode..."
if [ -x "$(command -v apt-get)" ]; then
    apt-get update && apt-get install -y dmidecode
elif [ -x "$(command -v yum)" ]; then
    yum install -y dmidecode
elif [ -x "$(command -v dnf)" ]; then
    dnf install -y dmidecode
fi

echo "Installation complete. The agent is now running."