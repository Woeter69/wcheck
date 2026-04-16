# wcheck - High-performance Headless Web Scanner

[![Go Test](https://github.com/Woeter69/wcheck/actions/workflows/test.yml/badge.svg)](https://github.com/Woeter69/wcheck/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/Woeter69/wcheck)](https://goreportcard.com/report/github.com/Woeter69/wcheck)

`wcheck` is a high-performance CLI tool written in Go designed to stress-test and audit your web applications using a headless browser. It simulates real user behavior to catch bugs that traditional crawlers miss, such as JavaScript crashes, broken interactions, and hidden network failures. It is especially useful for scanning localhost websites during development.

## 🚀 Features

- **Headless Browser Engine:** Uses real Chrome/Chromium via `chromedp` to accurately execute JavaScript and render pages.
- **Concurrent Scanning:** Utilizes a highly efficient worker pool pattern for high-speed, parallel audits of multiple internal pages.
- **Automated Link Discovery:** Automatically crawls and maps internal links on your domain to build a comprehensive scan queue.
- **Deep Error Sniffing:**
  - `runtime.EventExceptionThrown`: Catches unhandled JS crashes.
  - `runtime.EventConsoleAPICalled`: Captures `console.error()` outputs.
  - `network.EventLoadingFailed`: Detects 404s/500s for static assets and API calls.
- **Smarter Interaction Monkey:** A coordinate-based clicking mechanism that bypasses transparent overlays, waits for hydration, and simulates clicks on interactive elements (buttons, links, inputs) to trigger state-dependent bugs.
- **Actionable Reporting:** Detailed, color-coded CLI reports complete with stack traces and line numbers.

## 🧠 How it Works

`wcheck` operates in three distinct phases:

1.  **Scout Phase:** The tool navigates to the target URL, parses the DOM, and discovers all internal links to build a comprehensive scan queue.
2.  **Scan Phase:** Multiple parallel workers visit each discovered page, monitoring for JS exceptions, console errors, and 4xx/5xx network failures.
3.  **Interact Phase (Smarter Monkey):** For each page, the "Monkey" identifies all interactive elements and simulates user clicks to uncover hidden, state-dependent bugs.

## 📦 Installation

`wcheck` requires **Go 1.23+** and a **Chromium-based browser** (Chrome, Chromium, Brave) installed on your system.

### Quick Install (From Source)
```bash
git clone https://github.com/Woeter69/wcheck.git
cd wcheck
make build
sudo make install
```

For detailed installation instructions per OS (Arch Linux, Ubuntu/Debian), see [INSTALL.md](./INSTALL.md).

## 💻 Usage

`wcheck` provides two main commands: `scan` for crawling/bulk auditing and `interact` for deep-diving into a single page.

### Bulk Auditing (`scan`)
The primary command to crawl internal links and visit each page using parallel workers.

```bash
# Basic scan of a local server
wcheck scan http://localhost:3000

# Scan with 10 parallel workers and the interaction monkey enabled
wcheck scan http://localhost:3000 --workers 10 --interact

# Scan with custom timeout (default 30s)
wcheck scan http://localhost:3000 --timeout 60
```

### Targeted Testing (`interact`)
Targets a single page for heavy interaction testing. Ideal for debugging state-heavy React/Next.js pages.

```bash
# Interact with a specific page with verbose logging and up to 50 simulated clicks
wcheck interact http://localhost:3000/dashboard -v --max-clicks 50
```

### Key Flags
- `-w, --workers`: Number of parallel browser workers (default: 5). Increase for faster scans, but beware of overloading your local server.
- `-d, --delay`: Delay in seconds between page visits (default: 0). Useful for bypassing rate limits.
- `-i, --interact`: Enables the "Smarter Monkey" interaction phase.
- `-v, --verbose`: Enables detailed debug logging (coordinates, network events).

For more examples, see [USAGE.md](./USAGE.md).

## 🤖 CI/CD Integration

`wcheck` uses standard exit codes. You can easily integrate it into your CI/CD pipelines to fail builds if JS errors or broken links are found:

```bash
wcheck scan http://localhost:3000 -i || exit 1
```

## 🤝 Contributing

Contributions are welcome! Whether it's reporting a bug, suggesting a feature, or submitting a Pull Request.

### Development Setup
1. Fork and clone the repository.
2. Ensure you have Go 1.23+ installed.
3. Run tests before submitting: `go test ./... -v`

### Project Structure
- `cmd/`: CLI command definitions (Cobra).
- `internal/engine/`: Core headless browser logic and interaction monkey.
- `internal/crawler/`: Link discovery and site scouting.
- `internal/reporter/`: Pretty CLI reporting and table generation.
- `tests/`: Integration tests for scanner logic.

For more details, please read [CONTRIBUTING.md](./CONTRIBUTING.md).

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
