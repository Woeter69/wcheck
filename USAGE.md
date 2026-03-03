# Usage Guide

`wcheck` provides two main commands: `scan` for crawling/bulk auditing and `interact` for deep-diving into a single page.

## Core Commands

### `wcheck scan <URL>`
The primary audit command. It crawls internal links and visits each page using parallel workers.

```bash
# Basic scan
wcheck scan http://localhost:3000

# Scan with parallel workers and interaction monkey enabled
wcheck scan http://localhost:3000 --workers 10 --interact

# Scan with custom timeout (default 30s)
wcheck scan http://localhost:3000 --timeout 60
```

### `wcheck interact <URL>`
Targets a single page for heavy interaction testing. Ideal for debugging state-heavy React/Next.js pages.

```bash
# Interact with a specific page (dashboard) with verbose logging
wcheck interact http://localhost:3000/dashboard -v --max-clicks 50
```

## Important Flags

### `--workers` | `-w`
Sets the number of parallel browser workers.
- **Default:** 5
- **Advice:** Increase for faster bulk scans (e.g., 10-20), but avoid overloading local development servers.

### `--delay` | `-d`
Adds a wait time (in seconds) between each page visit.
- **Default:** 0
- **Advice:** Useful for avoiding rate limits (429 Too Many Requests) on production sites.

### `--interact` | `-i`
Enables the "Smarter Monkey" interaction phase on every discovered page.

### `--verbose` | `-v`
Enables detailed debug logging, showing coordinates for clicks and real-time capture of network/JS events.

## CI/CD Integration

Use the exit code to fail builds if errors are found:

```bash
wcheck scan http://localhost:3000 -i || exit 1
```
