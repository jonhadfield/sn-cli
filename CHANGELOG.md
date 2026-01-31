# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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