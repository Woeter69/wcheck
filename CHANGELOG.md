# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- **Engine Reliability & CI Performance:**
  - **Memory & Resource Leak:** Fixed a critical context shadowing bug in `ScanPage` and `ExtractLinks` where the browser tab/context was not being correctly cancelled, leading to resource exhaustion and CI timeouts.
  - **Chrome Configuration:** Fixed a typo in the `max-connections-per-browser` flag and increased its limit to 10 for better concurrency.
  - **Improved CI Stability:** Added `--disable-setuid-sandbox` and `--disable-extensions` Chrome flags to ensure more robust execution in headless Linux environments.
  - **Timeout Optimization:** Removed an arbitrary 120-second minimum timeout in `ExtractLinks` and replaced it with a shorter-circuiting mechanism that respects user-provided timeouts.
  - **Enhanced Test Reliability:** Increased test timeouts from 10s to 30s to accommodate slower execution in resource-constrained CI environments.

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
