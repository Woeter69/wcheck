# wcheck - High-performance Headless Web Scanner

wcheck is a professional CLI tool designed to stress-test and audit your web applications using a headless browser. It simulates real user behavior to catch bugs that traditional crawlers miss.

## What is it?
`wcheck` is a Go-based scanner that leverages `chromedp` (Chrome DevTools Protocol) to navigate your website. Unlike simple HTTP clients, `wcheck` executes JavaScript, allowing it to:
- Detect **Runtime JS Exceptions**.
- Catch **Console Errors** and warnings.
- Identify **Network Failures** (404s, 500s) for dynamically loaded assets.
- Validate complex user interactions.

## Key Features
- **Fast Link Discovery:** Automatically crawls your domain to build a complete map of internal pages.
- **Concurrent Worker Pool:** Scans multiple pages in parallel using a configurable worker architecture.
- **Interactive Monkey Testing:** Automatically finds and clicks interactive elements (buttons, links) to trigger hidden state bugs.
- **Pretty Terminal Reporting:** Clean, high-signal output with detailed summaries and execution timing.

## Quick Start
Scan your local development server with interaction testing enabled:
```bash
wcheck scan http://localhost:3000 -i
```

For more details, see [INSTALL.md](INSTALL.md) and [USAGE.md](USAGE.md).
