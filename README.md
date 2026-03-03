# wcheck - High-performance Headless Web Scanner

[![Go Test](https://github.com/Woeter69/wcheck/actions/workflows/test.yml/badge.svg)](https://github.com/Woeter69/wcheck/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/Woeter69/wcheck)](https://goreportcard.com/report/github.com/Woeter69/wcheck)

`wcheck` is a professional CLI tool designed to stress-test and audit your web applications using a headless browser. It simulates real user behavior to catch bugs that traditional crawlers miss, such as JavaScript crashes, broken interactions, and hidden network failures.

## How it Works

`wcheck` operates in three distinct phases:

1.  **Scout Phase:** The tool navigates to the target URL and discovers all internal links to build a scan queue.
2.  **Scan Phase:** Multiple parallel workers visit each discovered page, monitoring for JS exceptions, console errors, and 4xx/5xx network failures.
3.  **Interact Phase (Smarter Monkey):** For each page, the "Monkey" identifies all interactive elements (buttons, links, inputs) and simulates clicks to trigger state-dependent bugs.

## Features

- **Headless Browser:** Uses real Chrome/Chromium via `chromedp` to execute JavaScript.
- **Concurrent Scanning:** Worker pool pattern for high-speed audits.
- **Smarter Interaction Monkey:** Coordinate-based clicking that bypasses transparent overlays and waits for hydration.
- **Deep Dive Reporting:** Detailed reports with stack traces, line numbers, and color-coded error categories.

## Quick Start

```bash
# Install (requires Go 1.23+)
make build
sudo make install

# Scan a website with interaction testing
wcheck scan https://example.com -i
```

See [INSTALL.md](./INSTALL.md) for detailed setup and [USAGE.md](./USAGE.md) for command examples.
