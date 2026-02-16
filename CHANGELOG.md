# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- The Deep Fix: Thread Safety & Context Cancellation:
  - Wrapped error reporting in `sync.Mutex` to prevent race conditions.
  - Implemented explicit `page.EventLoadEventFired` synchronization using a channel and `chromedp.ActionFunc`.
  - Added immediate `[DEBUG]` terminal output when listeners catch an error (requires `--verbose`).
  - Added `disable-dev-shm-usage` flag to the browser allocator for improved headless stability on Linux.
  - Explicitly enabled the `Page` domain in `chromedp`.
- The Fix: Debugging the "Sniffer":
  - Ensured `chromedp.ListenTarget` is called before navigation.
  - Added `runtime.Enable()` to catch more asynchronous JavaScript errors.
  - Increased settle period to 2 seconds to allow background errors to trigger.
  - Improved console error capturing to include all message arguments.
  - Switched to `chromedp.WaitVisible` for body rendering confirmation.
  - Added cancellation check for network failures to reduce noise.

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
