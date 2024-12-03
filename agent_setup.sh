#!/bin/bash

# Required values
AUTH_TOKEN="" # validation token when sending collected data
URL="" # URL to send collected data
BINARY_URL="" # URL where binary is stored
# Optional values
UNINSTALL=false
COLLECT_INTERVAL_SEC="5" # Collect data every...
SEND_INTERVAL_SEC="15" # Send collected data every...
# Constants
AGENT_BINARY="/usr/local/bin/uptinio-agent"
SERVICE_NAME="uptinio-agent.service"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME"

# Helper function to validate URL format
validate_url_format() {
    local url=$1
    if [[ ! $url =~ ^http(s)?://[a-zA-Z0-9.-]+(:[0-9]+)?(/.*)?$ ]]; then
        echo "Error: Invalid URL format: $url"
        exit 1
    fi
}

# Helper function to check if URL is accessible
check_url_accessibility() {
    local url=$1
    if ! curl -Is "$url" >/dev/null 2>&1; then
        echo "Error: URL is not accessible: $url"
        exit 1
    fi
}

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --auth-token) AUTH_TOKEN="$2"; shift ;;
        --url) URL="$2"; shift ;;
        --binary-url) BINARY_URL="$2"; shift ;;
        --collect-interval-sec) COLLECT_INTERVAL_SEC="$2"; shift ;;
        --send-interval-sec) SEND_INTERVAL_SEC="$2"; shift ;;
        --uninstall) UNINSTALL=true ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

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

if [ -z "$URL" ]; then
    echo "Error: --url is required."
    exit 1
fi

if [ -z "$BINARY_URL" ]; then
    echo "Error: --binary-url is required."
    exit 1
fi

# Validate URL formats
validate_url_format "$URL"
validate_url_format "$BINARY_URL"

# Check if URLs are accessible
check_url_accessibility "$URL"
check_url_accessibility "$BINARY_URL"

# Install agent binary
echo "Downloading and installing agent binary..."
curl -Lo "$AGENT_BINARY" "$BINARY_URL"
chmod +x "$AGENT_BINARY"
echo "Binary installed at $AGENT_BINARY"

# Run the binary to create the configuration file
echo "Creating configuration using the binary..."
"$AGENT_BINARY" --create-config \
  --auth-token "$AUTH_TOKEN" \
  --url "$URL" \
  --collect-interval-sec "$COLLECT_INTERVAL_SEC" \
  --send-interval-sec "$SEND_INTERVAL_SEC"

# Create systemd service file
echo "Creating systemd service file..."
cat <<EOF >"$SERVICE_FILE"
[Unit]
Description=Uptinio Server Monitoring Agent
After=network.target

[Service]
ExecStart=$AGENT_BINARY
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

echo "Installation complete. The agent is now running."



# curl -fsSL https://your-server.com/install.sh | bash -s -- --auth-token "my-secret-token" --url "http://my-server.com/api/v1/server_metrics" --collect-interval-sec 10 --send-interval-sec 60