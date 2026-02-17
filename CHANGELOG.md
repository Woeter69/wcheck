# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Final Polish (Part 5):
  - Integrated `pterm` for a professional summary table.
  - Added total scan time tracking.
  - Implemented proper exit codes (1 on failure, 0 on success).
  - Standardized logging with a structured reporter.
- Implemented Worker Pool Pattern (Part 4) for concurrent scanning.
- Spawned multiple worker goroutines based on the `--workers` flag.
- Integrated `sync.WaitGroup` and channels for thread-safe job orchestration.
- Added per-worker debug logging.
- Implemented Link Discovery (Part 3).
- Added `ExtractLinks` method to `Engine` to retrieve anchor tags from pages.
- Added `Crawler` logic to resolve relative URLs, filter external domains, and deduplicate links.
- Added verbose output for discovered links.

### Changed
- Refactored the project structure to better separate concerns.
- Moved browser logic to `internal/engine`.
- Moved output formatting to `internal/reporter`.
- Introduced `internal/crawler` for link discovery logic.
- Consolidated Cobra setup into `main.go`.
- Updated `cmd/scan.go` to use the new internal packages.
