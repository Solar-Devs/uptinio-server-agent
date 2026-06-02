#!/bin/sh
# Run inside a Linux container (Debian, Ubuntu, CentOS, Fedora, Alpine, Arch, etc.).
set -eu

GO_VERSION="${GO_VERSION:-1.23.4}"
ARCH="${ARCH:-amd64}"

install_go() {
	if command -v go >/dev/null 2>&1; then
		current="$(go env GOVERSION 2>/dev/null || true)"
		if [ "${current}" = "go${GO_VERSION}" ]; then
			return
		fi
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
		apt-get install -y -qq curl ca-certificates dmidecode procps
	elif [ -f /etc/alpine-release ]; then
		apk add --no-cache curl ca-certificates dmidecode procps
	elif [ -f /etc/arch-release ]; then
		pacman -Sy --noconfirm --needed curl ca-certificates dmidecode procps-ng
	elif [ -f /etc/fedora-release ] || [ -f /etc/redhat-release ]; then
		rhel_pkgs="ca-certificates dmidecode procps-ng"
		if ! command -v curl >/dev/null 2>&1; then
			rhel_pkgs="curl ${rhel_pkgs}"
		fi
		if command -v dnf >/dev/null 2>&1; then
			dnf install -y ${rhel_pkgs}
		elif command -v yum >/dev/null 2>&1; then
			yum install -y ${rhel_pkgs}
		fi
	else
		echo "Unknown distro; continuing without extra packages"
	fi
}

install_distro_deps
install_go

cd /src
export CGO_ENABLED=0
export PATH="/usr/local/go/bin:${PATH}"
go test ./... -count=1 -v
