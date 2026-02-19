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
- **Smarter Interaction Monkey:**
  - **Refined Selection:** Now only targets likely interactive elements (`button`, `a`, `.btn`, etc.) and ignores decorative `svg` or `span` elements.
  - **Deduplication:** Automatically skips redundant elements with the same text (e.g., duplicate mobile/desktop menus).
  - **Pre-Click Validation:** Uses JavaScript to verify `cursor: pointer` and element visibility before attempting interaction.
  - **Performance Optimization:** Tightened per-click timeouts to 3 seconds and reduced post-click wait times for faster sessions.
  - **Reduced Noise:** Removed repetitive failure logs, focusing only on real caught exceptions.

- **Rate-Limiting Resilience:**
  - **Worker Delay:** Added `--delay` (or `-d`) flag to `scan` and `interact` commands to introduce a sleep between page scans, avoiding rate limits.
  - **Retry Logic:** Workers now automatically retry a page scan once after a 5-second sleep if they encounter a timeout (`context.DeadlineExceeded`) or a `429 Too Many Requests` error.
  - **Connection Limits:** Restricted Chrome to 2 connections per browser via `max-connection-per-browser` flag for reduced host impact.
  - **Faster Extraction:** Switched from `WaitVisible` to `WaitReady` for link discovery, returning as soon as the DOM is ready.
  - **429 Detection:** The network listener now explicitly identifies and reports `429` status codes.

- **The Interaction Monkey (`interact` command):**
  - New `wcheck interact <URL>` command for targeted interaction testing.
  - **Discovery Engine:** Automatically finds all `button`, `a`, `[role="button"]`, and elements with `cursor: pointer` via JS evaluation.
  - **The "Monkey" Logic:**
    - Scrolls elements into view before interacting.
    - Performs automated `chromedp.Click`.
    - **Safety Filter:** Blacklists sensitive keywords (`logout`, `delete`, `remove`, `sign out`) to prevent accidental session termination or data loss.
    - **State Reset:** Automatically navigates back to the original URL after each click to ensure state consistency.
    - **Limit Control:** Added `--max-clicks` flag (default 20) to prevent infinite loops on complex pages.
  - **Detailed Interaction Reporting:**
    - New `FAILED INTERACTIONS` section in the reporter.
    - Captures exactly which element (by Text, ID, or XPath) caused a crash.
    - Re-categorizes JS exceptions and console errors as "Broken Interaction" when triggered during the interaction phase.

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
