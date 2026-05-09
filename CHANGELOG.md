# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.2] - 2026-05-08

### Added
- SSH keepalive options (`ServerAliveInterval=60`, `ServerAliveCountMax=3`, `TCPKeepAlive=yes`) to prevent silent connection drops
- Auto-reconnect functionality with configurable interval (default: 30 seconds)
- `reconnect_interval` configuration option per profile (0 to disable)
- Background reconnect monitor that checks tunnel health and restarts dead tunnels
- Signal handling (SIGINT/SIGTERM) for graceful shutdown
- `portbridge version` command

## [0.0.1] - 2026-05-08

### Added
- Initial release
- Profile-based SSH tunnel management
- One-command tunnel startup
- Enable/disable tunnels per profile
- YAML-based configuration
