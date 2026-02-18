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
- **Detailed Error Reporting & Deep Dive:**
  - Implemented `PageError` structure to capture JS stack traces, line numbers, and error types.
  - Added a "DEEP DIVE: ERROR DETAILS" section to the final report with color-coded errors (Red for JS, Yellow for Console, Cyan for Network).
  - Enabled **Asynchronous Stack Traces** via the Chrome debugger domain to track errors across `setTimeout` and promises.
- **Flexible Timeout Control:**
  - Added `--timeout` (or `-t`) flag to the `scan` command for per-page load control.
  - Implemented graceful timeout reporting (`⌛ Page load timed out after [N]s`).
- **Resilient Link Discovery (Scout):**
  - Updated link extraction to use a 45s minimum timeout regardless of global settings.
  - Implemented a fallback strategy: if a page body fails to appear, the crawler still attempts to extract links from the partial DOM.
  - Added verbose logging for the "Scout" phase (`Scout is waiting for body to appear...`).

### Changed
- **Smarter Waiting Strategy:**
  - Replaced rigid `time.Sleep` with `chromedp.WaitReady` and `WaitVisible` for more efficient page settling.
  - Refined console error extraction to remove unnecessary quotes and improve readability.
  - Updated `GetLinks` and `ScanPage` signatures to support dynamic timeouts.

### Fixed
- Fixed JS exception messages that were duplicating stack traces in the description.
- Resolved "context deadline exceeded" errors on heavy pages by decoupling crawler and worker timeouts.

### Changed
- Refactored the project structure to better separate concerns.
- Moved browser logic to `internal/engine`.
- Moved output formatting to `internal/reporter`.
- Introduced `internal/crawler` for link discovery logic.
- Consolidated Cobra setup into `main.go`.
- Updated `cmd/scan.go` to use the new internal packages.
