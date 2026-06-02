#!/bin/bash
#
# Uptinio agent installer — supports Debian, Ubuntu, CentOS, Fedora, Alpine, and Arch.
# See AGENTS.md for distro compatibility expectations.

set -euo pipefail

if [[ $EUID -ne 0 ]]; then
  echo "This script must be run with sudo or as root." >&2
  exit 1
fi

# Required values
AUTH_TOKEN=""
# Optional values
HOST="app.uptinio.com"
SCHEMA=https
COLLECT_INTERVAL=60
SEND_INTERVAL=60
METRICS_PATH=/var/tmp/uptinio-agent/metrics.json
LOG_PATH=/var/log/uptinio-agent/agent.log
MAX_LOG_SIZE=1024
CONFIG_PATH=/etc/uptinio-agent.yaml
UNINSTALL=false
# Set LOCAL_AGENT_BINARY to a local file path to skip downloading (used in CI).
LOCAL_AGENT_BINARY="${LOCAL_AGENT_BINARY:-}"

# Constants (Alpine/musl uses a separate release artifact)
BINARY_URL_GLIBC=https://github.com/Solar-Devs/uptinio-server-agent/releases/latest/download/agent-linux-amd64
BINARY_URL_MUSL=https://github.com/Solar-Devs/uptinio-server-agent/releases/latest/download/agent-linux-musl-amd64
AGENT_BINARY=/usr/local/bin/uptinio-agent
SYSTEMD_SERVICE_NAME=uptinio-agent.service
SYSTEMD_SERVICE_FILE=/etc/systemd/system/$SYSTEMD_SERVICE_NAME
INITD_SERVICE_NAME=uptinio-agent
INITD_SCRIPT=/etc/init.d/$INITD_SERVICE_NAME
OPENRC_SCRIPT=/etc/init.d/$INITD_SERVICE_NAME
PID_FILE=/var/run/uptinio-agent.pid

detect_distro_family() {
  if [ -f /etc/debian_version ]; then
    echo "debian"
  elif [ -f /etc/alpine-release ]; then
    echo "alpine"
  elif [ -f /etc/arch-release ]; then
    echo "arch"
  elif [ -f /etc/fedora-release ]; then
    echo "fedora"
  elif [ -f /etc/redhat-release ] || [ -f /etc/centos-release ] || [ -f /etc/rocky-release ]; then
    echo "rhel"
  else
    echo "unknown"
  fi
}

run_pkg_install() {
  local pkgs=("$@")
  local family
  family=$(detect_distro_family)

  case "$family" in
    debian)
      export DEBIAN_FRONTEND=noninteractive
      apt-get update -qq
      apt-get install -y -qq "${pkgs[@]}"
      ;;
    alpine)
      apk add --no-cache "${pkgs[@]}"
      ;;
    arch)
      pacman -Sy --noconfirm --needed "${pkgs[@]}"
      ;;
    fedora|rhel)
      local install_pkgs=()
      local pkg
      for pkg in "${pkgs[@]}"; do
        # CentOS Stream ships curl-minimal; installing curl conflicts.
        if [ "$pkg" = "curl" ] && command -v curl >/dev/null 2>&1; then
          continue
        fi
        install_pkgs+=("$pkg")
      done
      if [ "${#install_pkgs[@]}" -eq 0 ]; then
        return 0
      fi
      if command -v dnf >/dev/null 2>&1; then
        dnf install -y "${install_pkgs[@]}"
      elif command -v yum >/dev/null 2>&1; then
        yum install -y "${install_pkgs[@]}"
      else
        echo "Warning: no dnf/yum found; could not install: ${install_pkgs[*]}" >&2
        return 1
      fi
      ;;
    *)
      echo "Warning: unknown distro family; could not install: ${pkgs[*]}" >&2
      return 1
      ;;
  esac
}

ensure_download_tool() {
  if command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1; then
    return 0
  fi
  echo "Installing curl..."
  run_pkg_install curl ca-certificates
}

install_dmidecode() {
  echo "Installing dmidecode (if available)..."
  if ! run_pkg_install dmidecode; then
    echo "Warning: dmidecode could not be installed; motherboard ID may use fallback." >&2
  fi
}

agent_binary_url() {
  if [ "$(detect_distro_family)" = "alpine" ]; then
    echo "$BINARY_URL_MUSL"
  else
    echo "$BINARY_URL_GLIBC"
  fi
}

download_agent_binary() {
  if [ -n "$LOCAL_AGENT_BINARY" ] && [ -f "$LOCAL_AGENT_BINARY" ]; then
    echo "Installing agent binary from $LOCAL_AGENT_BINARY..."
    install -m 755 "$LOCAL_AGENT_BINARY" "$AGENT_BINARY"
    return 0
  fi

  ensure_download_tool
  local url
  url=$(agent_binary_url)
  echo "Downloading and installing agent binary from $url ..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$AGENT_BINARY" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$AGENT_BINARY" "$url"
  else
    echo "Error: curl or wget is required to download the agent." >&2
    exit 1
  fi
  chmod +x "$AGENT_BINARY"
}

has_systemctl() {
  command -v systemctl >/dev/null 2>&1 && systemctl list-units --type=service >/dev/null 2>&1
}

has_openrc() {
  command -v openrc-run >/dev/null 2>&1 && command -v rc-service >/dev/null 2>&1
}

is_openrc_service() {
  [ -f "$OPENRC_SCRIPT" ] && head -n 1 "$OPENRC_SCRIPT" | grep -q openrc-run
}

run_initd() {
  local action=$1
  if command -v rc-service >/dev/null 2>&1 && is_openrc_service; then
    rc-service "$INITD_SERVICE_NAME" "$action"
  elif command -v service >/dev/null 2>&1; then
    service "$INITD_SERVICE_NAME" "$action"
  elif [ -x "$INITD_SCRIPT" ]; then
    "$INITD_SCRIPT" "$action"
  else
    echo "Error: cannot $action $INITD_SERVICE_NAME (no service manager found)." >&2
    return 1
  fi
}

install_openrc_service() {
  echo "Creating OpenRC service: $OPENRC_SCRIPT..."
  mkdir -p "$(dirname "$METRICS_PATH")" "$(dirname "$LOG_PATH")" "$(dirname "$OPENRC_SCRIPT")"
  cat <<EOF >"$OPENRC_SCRIPT"
#!/sbin/openrc-run

name="$INITD_SERVICE_NAME"
description="Uptinio Server Monitoring Agent"
command="$AGENT_BINARY"
command_args="--config-path $CONFIG_PATH"
command_background="yes"
pidfile="$PID_FILE"

depend() {
    need net
    after networking
}
EOF
  chmod +x "$OPENRC_SCRIPT"
  rc-update add "$INITD_SERVICE_NAME" default
  rc-service "$INITD_SERVICE_NAME" start 2>/dev/null || rc-service "$INITD_SERVICE_NAME" restart 2>/dev/null || true
}

install_sysv_service() {
  echo "Creating SysV init script: $INITD_SCRIPT..."
  mkdir -p "$(dirname "$INITD_SCRIPT")"
  cat <<EOF >"$INITD_SCRIPT"
#!/bin/sh
### BEGIN INIT INFO
# Provides:          $INITD_SERVICE_NAME
# Required-Start:    \$network \$remote_fs
# Required-Stop:     \$network \$remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Uptinio Server Monitoring Agent
### END INIT INFO

AGENT_BINARY="$AGENT_BINARY"
CONFIG_PATH="$CONFIG_PATH"
PIDFILE="$PID_FILE"
METRICS_PATH="$METRICS_PATH"
LOG_PATH="$LOG_PATH"

start() {
    if [ -f "\$PIDFILE" ] && kill -0 "\$(cat "\$PIDFILE")" 2>/dev/null; then
        echo "$INITD_SERVICE_NAME is already running"
        return 0
    fi
    mkdir -p "\$(dirname "\$METRICS_PATH")" "\$(dirname "\$LOG_PATH")"
    if command -v start-stop-daemon >/dev/null 2>&1; then
        start-stop-daemon --start --background --pidfile "\$PIDFILE" --make-pidfile \\
            --exec "\$AGENT_BINARY" -- --config-path "\$CONFIG_PATH"
    else
        "\$AGENT_BINARY" --config-path "\$CONFIG_PATH" >>"\$LOG_PATH" 2>&1 &
        echo \$! >"\$PIDFILE"
    fi
    echo "Started $INITD_SERVICE_NAME"
}

stop() {
    if command -v start-stop-daemon >/dev/null 2>&1 && [ -f "\$PIDFILE" ]; then
        start-stop-daemon --stop --pidfile "\$PIDFILE" --retry 5 2>/dev/null || true
    elif [ -f "\$PIDFILE" ]; then
        pid=\$(cat "\$PIDFILE" 2>/dev/null)
        if [ -n "\$pid" ] && kill -0 "\$pid" 2>/dev/null; then
            kill "\$pid" 2>/dev/null
            sleep 1
            kill -0 "\$pid" 2>/dev/null && kill -9 "\$pid" 2>/dev/null
        fi
    fi
    rm -f "\$PIDFILE"
    pkill -x uptinio-agent 2>/dev/null || true
    echo "Stopped $INITD_SERVICE_NAME"
}

status() {
    if [ -f "\$PIDFILE" ] && kill -0 "\$(cat "\$PIDFILE")" 2>/dev/null; then
        echo "$INITD_SERVICE_NAME is running (PID \$(cat "\$PIDFILE"))"
        return 0
    fi
    echo "$INITD_SERVICE_NAME is not running"
    return 1
}

case "\$1" in
    start)   start ;;
    stop)    stop ;;
    restart) stop; start ;;
    status)  status ;;
    *)       echo "Usage: \$0 {start|stop|restart|status}"; exit 1 ;;
esac
EOF
  chmod +x "$INITD_SCRIPT"

  if command -v update-rc.d >/dev/null 2>&1; then
    update-rc.d "$INITD_SERVICE_NAME" defaults
  elif command -v chkconfig >/dev/null 2>&1; then
    chkconfig --add "$INITD_SERVICE_NAME"
    chkconfig "$INITD_SERVICE_NAME" on
  elif command -v rc-update >/dev/null 2>&1; then
    rc-update add "$INITD_SERVICE_NAME" default
  else
    echo "Warning: could not enable boot startup (update-rc.d/chkconfig/rc-update not found)."
  fi

  run_initd start
}

install_systemd_service() {
  echo "Creating systemd service file..."
  cat <<EOF >"$SYSTEMD_SERVICE_FILE"
[Unit]
Description=Uptinio Server Monitoring Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$AGENT_BINARY --config-path $CONFIG_PATH
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable "$SYSTEMD_SERVICE_NAME"
  systemctl start "$SYSTEMD_SERVICE_NAME"
}

install_service() {
  mkdir -p "$(dirname "$METRICS_PATH")" "$(dirname "$LOG_PATH")"

  if has_systemctl; then
    install_systemd_service
    echo "service-manager=systemd" > /etc/uptinio-agent.service-manager
  elif has_openrc; then
    install_openrc_service
    echo "service-manager=openrc" > /etc/uptinio-agent.service-manager
  else
    install_sysv_service
    echo "service-manager=sysv" > /etc/uptinio-agent.service-manager
  fi
}

remove_systemd_service() {
  if systemctl is-active --quiet "$SYSTEMD_SERVICE_NAME" 2>/dev/null; then
    echo "Stopping systemd service: $SYSTEMD_SERVICE_NAME"
    systemctl stop "$SYSTEMD_SERVICE_NAME"
  fi
  if systemctl is-enabled --quiet "$SYSTEMD_SERVICE_NAME" 2>/dev/null; then
    echo "Disabling systemd service: $SYSTEMD_SERVICE_NAME"
    systemctl disable "$SYSTEMD_SERVICE_NAME"
  fi
  if [ -f "$SYSTEMD_SERVICE_FILE" ]; then
    echo "Removing systemd service file: $SYSTEMD_SERVICE_FILE"
    rm -f "$SYSTEMD_SERVICE_FILE"
  fi
  systemctl daemon-reload 2>/dev/null || true
}

remove_openrc_service() {
  if [ -f "$OPENRC_SCRIPT" ] && is_openrc_service; then
    rc-service "$INITD_SERVICE_NAME" stop 2>/dev/null || true
    rc-update del "$INITD_SERVICE_NAME" default 2>/dev/null || true
    echo "Removing OpenRC script: $OPENRC_SCRIPT"
    rm -f "$OPENRC_SCRIPT"
  fi
}

remove_sysv_service() {
  if [ -f "$INITD_SCRIPT" ] && ! is_openrc_service; then
    run_initd stop 2>/dev/null || true
    if command -v update-rc.d >/dev/null 2>&1; then
      update-rc.d "$INITD_SERVICE_NAME" remove
    elif command -v chkconfig >/dev/null 2>&1; then
      chkconfig --del "$INITD_SERVICE_NAME" 2>/dev/null || true
    fi
    echo "Removing init script: $INITD_SCRIPT"
    rm -f "$INITD_SCRIPT"
  fi
}

stop_agent_process() {
  pkill -x uptinio-agent 2>/dev/null || true
  rm -f "$PID_FILE"
}

remove_service() {
  local manager=""
  if [ -f /etc/uptinio-agent.service-manager ]; then
    manager=$(cat /etc/uptinio-agent.service-manager 2>/dev/null || true)
  fi

  case "$manager" in
    systemd) remove_systemd_service ;;
    openrc)  remove_openrc_service ;;
    sysv)    remove_sysv_service ;;
    *)
      remove_systemd_service
      remove_openrc_service
      remove_sysv_service
      ;;
  esac
  stop_agent_process
  rm -f /etc/uptinio-agent.service-manager
}

# Parse arguments
if [[ "$#" -eq 1 && "$1" != "--uninstall" ]]; then
  AUTH_TOKEN="$1"
else
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

if [ "$UNINSTALL" = "true" ]; then
  echo "Uninstalling uptinio-agent..."
  remove_service

  echo "Removing configuration file: $CONFIG_PATH"
  rm -f "$CONFIG_PATH"

  echo "Removing metrics file: $METRICS_PATH"
  rm -f "$METRICS_PATH"

  echo "Removing log file: $LOG_PATH"
  rm -f "$LOG_PATH"

  if [ -f "$AGENT_BINARY" ]; then
    echo "Removing binary: $AGENT_BINARY"
    rm -f "$AGENT_BINARY"
  fi

  echo "Uninstallation complete."
  exit 0
fi

if [ -z "$AUTH_TOKEN" ]; then
  echo "Error: --auth-token is required."
  exit 1
fi

DISTRO_FAMILY=$(detect_distro_family)
echo "Detected distro family: $DISTRO_FAMILY"

download_agent_binary
echo "Binary installed at $AGENT_BINARY"

echo "Creating configuration file: $CONFIG_PATH..."
cat <<EOF >"$CONFIG_PATH"
metrics_path: "$METRICS_PATH"
log_path: "$LOG_PATH"
max_log_file_size_in_MB: $MAX_LOG_SIZE
schema: "$SCHEMA"
host: "$HOST"
auth_token: "$AUTH_TOKEN"
collect_interval_in_seconds: $COLLECT_INTERVAL
send_interval_in_seconds: $SEND_INTERVAL
EOF

install_service
install_dmidecode

case "$(cat /etc/uptinio-agent.service-manager 2>/dev/null || echo unknown)" in
  systemd)
    echo "Installation complete. The agent is running (systemd: $SYSTEMD_SERVICE_NAME)."
    ;;
  openrc)
    echo "Installation complete. The agent is running (OpenRC: rc-service $INITD_SERVICE_NAME status)."
    ;;
  sysv)
    echo "Installation complete. The agent is running (SysV: service $INITD_SERVICE_NAME status)."
    ;;
  *)
    echo "Installation complete."
    ;;
esac
