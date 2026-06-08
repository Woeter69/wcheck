# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- **Engine & Interaction Reliability:**
  - **Error Deduplication:** Implemented a robust deduplication mechanism to prevent the same error from being reported multiple times.
  - **Precise Attribution:** Improved the logic for attributing errors during the interaction phase, ensuring initial page load errors are no longer incorrectly flagged as "Broken Interactions".
  - **Interaction Isolation:** Added state resets and settle times between clicks to prevent cross-contamination of errors between different interactive elements.
  - **Smarter Visibility Check:** Refined the `isVisible` logic to better handle modern CSS (opacity, offsetParent) and ensure only truly interactive elements are targeted.
  - **Network Error Sniffing:** Improved capturing of HTTP 4xx/5xx errors during both the crawl and interaction phases.

### Added
- **Professional Documentation Suite:**
  - `README.md`: New sleek project overview and feature highlights.
  - `INSTALL.md`: Detailed installation instructions and source builds.
  - `USAGE.md`: Comprehensive breakdown of all CLI flags and CI/CD examples.
  - `CONTRIBUTING.md`: Developer-focused guide on project structure.

## [1.0.1] - 2026-03-03

### Fixed
- **Truthful Interaction Reporting:** Fixed a major bug where interaction timeouts (context deadline exceeded) were being reported as "OK". These are now accurately captured as `TypeErrorInteraction` and mark the scan as **FAIL**.
- **Hydration & Settle Phase:** Improved stability on modern JS-heavy sites (React/Next.js) by implementing explicit `WaitVisible('body')` and settle periods before scanning or interacting.
- **Selector Expansion:** Broadened the "Monkey" selector to include `.btn`, `.button`, and `input[type='submit']`, increasing element discovery by up to 300% on complex pages.
- **Coordinate-Based Clicking:** Switched to `MouseClickXY` to bypass transparent overlays that were blocking standard `chromedp.Click` actions.
- **Memory Safety:** Fixed a concurrent map write in the error listener.

### Added
- **Integration Test Suite:** Added a dedicated `tests/` folder with `httptest`-based integration tests to verify element discovery and timeout handling.
- **Version Flag:** Added `-V` / `--version` flag to the root command.
- **Comprehensive Documentation:** Added detailed README, INSTALL, USAGE, and CONTRIBUTING guides.
- **GitHub Actions:** CI workflow to run tests on every push.

## [1.0.0] - Initial Release
- **Core Functionality:** Scout, Scan, and Interact phases.
- **Headless Engine:** Built on `chromedp`.
- **Worker Pool:** Parallel page auditing.
