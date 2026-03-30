# Corridor CLI Changelog

## 1.4.2 (2026-03-30)

Added --verbose flag, improved install output, checksum verification fix.

- Add --verbose / -v flag and CORRIDOR_VERBOSE env var for controlling output levels
- Gate intermediate install/login output behind --verbose; keep essential output by default
- Fix checksum verification in install.sh to fail on mismatch instead of silently continuing
- Add --target flag to 'corridor install' for targeting specific agents

## 1.4.1 (2026-03-24)

Skip auto-update when running uninstall.

- Skip auto-update check for 'corridor uninstall' to avoid unnecessary downloads

## 1.4.0 (2026-03-20)

Self-update mechanism and plugin reinstall flow.

- Add auto-update: CLI checks for newer version before running commands
- Automatic plugin reinstall after self-update for compatibility
- New 'corridor update' command for manual updates

## 1.3.0 (2026-03-10)

Config management and status improvements.

- Add 'corridor config set/get' commands for managing config.env keys
- Add 'corridor status --json' for machine-readable status output
- Add 'corridor list' command to show installed plugins

## 1.2.0 (2026-02-28)

SSO login and multi-target support.

- Add SSO browser-based login flow via 'corridor login'
- Support multiple install targets (auto-detect or specify with --target)
- Write agent rules to target-specific directories on install

## 1.1.0 (2026-02-15)

Plugin system and uninstall support.

- Plugin-based architecture for managing tool integrations
- Add 'corridor uninstall' command to remove plugins
- Add 'corridor install --force' to reinstall existing plugins

## 1.0.0 (2026-02-01)

Initial release of Corridor CLI.

- Install corridor plugins for supported coding agents
- Platform detection for linux/darwin on amd64/arm64
- Installation script with checksum verification
