# wcheck - Headless Browser Scanner

## Project Overview
`wcheck` is a high-performance CLI tool written in Go for scanning localhost websites using a headless browser.

## Current Progress
- [x] Part 1: Initialize Go module and Cobra CLI (`scan` command with flags)
- [x] Part 2: Integrate `chromedp` and implement error listeners (JS Exceptions, Console Errors, Network Failures)
- [x] Part 3: Link Discovery (Internal page crawling)
- [x] Part 4: Worker Pool Implementation (Parallel Scanning)
- [x] Part 5: Final Polishing (Pretty reporting, exit codes, and timing)
- [x] Part 6: Installation & Automation (Makefile, Git Hooks)

## Features
- **Headless Browser:** Uses real Chrome/Chromium to execute JavaScript.
- **Concurrent Scanning:** Uses worker pool pattern to scan multiple internal pages.
- **Link Discovery:** Automatically crawls internal links on your domain.
- **Error Sniffing:**
  - `runtime.EventExceptionThrown`: Catch JS crashes.
  - `runtime.EventConsoleAPICalled`: Catch `console.error()`.
  - `network.EventLoadingFailed`: Catch 404s/500s for static assets.

## Command Usage
`wcheck scan <URL> -w <workers> -v`
