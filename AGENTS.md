# Agent guidelines (uptinio-server-agent)

Instructions for humans and coding agents working on this repository.

## Supported Linux distributions

All features, bug fixes, installer changes, and tests must be validated against these families:

| Distro   | CI image                         | Init (typical)     | Packages        |
|----------|----------------------------------|--------------------|-----------------|
| Debian   | `debian:12`                      | systemd / SysV     | `apt`           |
| Ubuntu   | `ubuntu:24.04`                   | systemd            | `apt`           |
| CentOS   | `quay.io/centos/centos:stream9`  | systemd / SysV     | `dnf` / `yum`   |
| Fedora   | `fedora:41`                      | systemd            | `dnf`           |
| Alpine   | `alpine:3.20`                    | **OpenRC**         | `apk`           |
| Arch     | `archlinux:latest`               | systemd            | `pacman`        |

Do not assume systemd-only, `apt`-only, or `top` output for metrics. Prefer portable approaches:

- **CPU metrics**: `gopsutil` / `/proc/stat` (not `top` parsing).
- **Service install** (`linux_installer.sh`): systemd → OpenRC → SysV, in that order.
- **Packages**: use `detect_distro_family` / `run_pkg_install` patterns from the installer or `scripts/ci-test-linux.sh`.

## Testing expectations

Before merging changes that touch metrics, install, or OS integration:

1. Run unit tests locally: `go test ./... -count=1`
2. CI runs `.github/workflows/test.yml`:
   - **unit** job on `ubuntu-latest`
   - **linux-distros** matrix (all six distros above) via `scripts/ci-test-linux.sh`
   - **installer-smoke** matrix via `scripts/ci-test-installer.sh`

When adding behavior, extend tests in the matching package and ensure the distro matrix still passes.

## Installer notes

- `linux_installer.sh` must remain **POSIX-friendly bash** and work as **root**.
- `LOCAL_AGENT_BINARY=/path/to/binary` skips the GitHub download (CI smoke tests).
- CentOS may ship `curl-minimal`; do not force-install the `curl` package if a curl binary already exists.
- Alpine requires an **OpenRC** init script (`#!/sbin/openrc-run`), not only SysV `update-rc.d`.
- Docker containers often lack a running systemd PID 1; the installer should still succeed using OpenRC or SysV.

## Payload and device identity

- Keep `motherboard_id` and other attributes bounded (see `SanitizeDeviceID`, 256-char cap).
- Avoid unbounded accumulation in the metrics file when sends fail (see `maxStoredMetrics` in `storage.go`).
- Handle HTTP **413** by trimming and retrying (see `sender.go`).

## Release binaries

Linux builds are static (`CGO_ENABLED=0`):

| Artifact                 | Targets                                      |
|--------------------------|----------------------------------------------|
| `agent-linux-amd64`      | Debian, Ubuntu, CentOS, Fedora, Arch (glibc) |
| `agent-linux-musl-amd64` | Alpine (musl) — **required** on Alpine         |

`linux_installer.sh` picks the correct URL via `agent_binary_url()`. Do not use the glibc binary on Alpine.

CI builds the musl artifact in `release.yml` using `golang:1.23-alpine`. Installer smoke tests build locally in each container via `LOCAL_AGENT_BINARY`.
