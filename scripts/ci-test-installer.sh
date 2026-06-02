#!/bin/sh
# Smoke-test linux_installer.sh inside the target distro container.
set -eu

GO_VERSION="${GO_VERSION:-1.23.4}"
ARCH="${ARCH:-amd64}"
AUTH_TOKEN="${AUTH_TOKEN:-ci-test-token}"

install_go() {
	need_install=1
	if command -v go >/dev/null 2>&1; then
		if go env GOVERSION 2>/dev/null | grep -qE 'go1\.(23|24|25)'; then
			need_install=0
		fi
	fi
	if [ "$need_install" -eq 0 ]; then
		return
	fi

	echo "Installing Go ${GO_VERSION}..."
	curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
		| tar -C /usr/local -xz
	export PATH="/usr/local/go/bin:${PATH}"
}

install_distro_deps() {
	if [ -f /etc/debian_version ]; then
		export DEBIAN_FRONTEND=noninteractive
		apt-get update -qq
		apt-get install -y -qq curl ca-certificates dmidecode procps bash
	elif [ -f /etc/alpine-release ]; then
		apk add --no-cache curl ca-certificates dmidecode procps bash openrc
	elif [ -f /etc/arch-release ]; then
		pacman -Sy --noconfirm --needed curl ca-certificates dmidecode procps-ng bash
	elif [ -f /etc/fedora-release ] || [ -f /etc/redhat-release ]; then
		rhel_pkgs="ca-certificates dmidecode procps-ng bash"
		if ! command -v curl >/dev/null 2>&1; then
			rhel_pkgs="curl ${rhel_pkgs}"
		fi
		if command -v dnf >/dev/null 2>&1; then
			dnf install -y ${rhel_pkgs}
		elif command -v yum >/dev/null 2>&1; then
			yum install -y ${rhel_pkgs}
		fi
	fi
}

install_distro_deps

if [ -f /etc/alpine-release ]; then
	mkdir -p /run/openrc
	touch /run/openrc/softlevel
	openrc 2>/dev/null || true
fi

install_go

cd /src
export CGO_ENABLED=0
export PATH="/usr/local/go/bin:${PATH}"

echo "Building agent binary..."
go build -o /tmp/uptinio-agent .

echo "Running installer smoke test..."
export LOCAL_AGENT_BINARY=/tmp/uptinio-agent
bash /src/linux_installer.sh --auth-token "$AUTH_TOKEN" --host 127.0.0.1 --schema http

test -x /usr/local/bin/uptinio-agent
test -f /etc/uptinio-agent.yaml
test -f /etc/uptinio-agent.service-manager

echo "Verifying agent process..."
sleep 3
if pgrep -x uptinio-agent >/dev/null 2>&1; then
	echo "Agent process is running."
elif command -v rc-service >/dev/null 2>&1 && rc-service uptinio-agent status 2>/dev/null | grep -Eq 'started|running'; then
	echo "Agent service is started (OpenRC)."
else
	echo "Error: uptinio-agent not running after install." >&2
	rc-service uptinio-agent status 2>&1 || true
	exit 1
fi

echo "Running uninstall..."
bash /src/linux_installer.sh --uninstall

if pgrep -x uptinio-agent >/dev/null 2>&1; then
	echo "Error: agent still running after uninstall." >&2
	exit 1
fi

test ! -f /usr/local/bin/uptinio-agent
echo "Installer smoke test passed."
