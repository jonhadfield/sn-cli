# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2026-09-21

### Added

- `--session-file`, `SN_SESSION_FILE` and `session_file` to store the session
  in a file instead of the system keyring, for headless servers with no
  keyring service (#94)
- Scheduled CI check that the Homebrew cask installs and runs

### Fixed

- Obsidian migration now implements the MOC styles it accepts, instead of
  ignoring them silently
- The fish completion script now works
- The install script no longer asks for sudo when `BIN_DIR` simply does not
  exist yet
- Refreshing an encrypted session no longer prompts for the session key again,
  which blocked unattended use

### Changed

- Build with Go 1.27.1. CI took the Go version from `go.mod`, so released
  binaries were missing later 1.26 standard-library fixes
- Updated dependencies, holding `google.golang.org/grpc` at v1.83.2 because
  v1.84.0 shipped without the fix for CVE-2026-84445
- Cleared all golangci-lint findings
- Corrected and restructured the README

## [0.5.3] - 2026-09-14

### Fixed

- The Homebrew cask no longer emits Homebrew's deprecated `postflight`

## [0.5.2] - 2026-09-13

### Added

- One-line install script that picks the right archive for the platform and
  verifies its checksum

## [0.5.1] - 2026-09-13

### Fixed

- An unknown subcommand now shows the help for the command it was given,
  rather than the top-level help

## [0.5.0] - 2026-09-13

### Added

- `sn editor` commands to manage the default note editor
- Homebrew cask published to jonhadfield/tap, with a release workflow

### Fixed

- Backup encryption, along with several SonarCloud findings
- `internal/sncli` tests failing to compile, and two broken unit tests
- Sync returning 401 in the `cmd/sncli` tests
- Broken install command in the README

### Changed

- Live-server tests are opt-in via `SN_INTEGRATION_TESTS`, so CI no longer
  touches a real account
- Removed hard-coded credentials from the test setup
- Dropped the unused password parameter from `GetBackupInfo`
- Updated dependencies and migrated to the supported Gemini SDK
- Fixed CI workflows, updated pinned actions, and reported every OS result

## [0.4.1] - 2026-01-30

### Fixed

- Authentication issues by updating gosn-v2 dependency to fix cookie-based auth
- Tag cloud to properly use Tag→Note references instead of Note→Tag
- Tag cloud to work completely offline using cached data
- Network error handling in tag cloud with graceful degradation
- Tag reference matching and display issues

### Improved

- Tag cloud now supports offline operation with cached items
- Enhanced debugging for note reference detection
- Better error messages for network failures

## [0.4.0] - 2026-01-29

### Added

- Backup and restore functionality with optional encryption
- Enhanced export with multiple format support (Markdown, HTML, static site)
- Tag cloud visualization for exploring note relationships
- Note templates system for quick note creation
- Visual improvements with better progress indicators

### Changed

- Updated authentication to use cache.GetSession for better session management

## [0.3.5] - 2024-01-08

### Fixed

- Fix conflict warning handling
- Minor code simplification

### Added

- Helper tests

## [0.3.4] - 2024-01-07

### Fixed

- Fix command completion and update instructions

## [0.3.3] - 2024-01-07

### Added

- Add `task` command for management of Checklists and Advanced Checklists

## [0.3.2] - 2024-01-06

### Fixed

- Bug fixes and sync speed increases

## [0.3.1] - 2023-12-20

### Improved

- Various output improvements, including stats

## [0.3.0] - 2023-12-14

### Fixed

- Bug fixes and item schema tests

## [0.2.8] - 2023-12-07

### Added

- Stored sessions are now auto-renewed when expired, or nearing expiry

## [0.2.7] - 2023-12-06

### Changed

- Various release packaging updates - thanks: [@clayrosenthal](https://github.com/clayrosenthal)
