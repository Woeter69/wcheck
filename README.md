# wcheck - Headless Browser Scanner

`wcheck` is a high-performance CLI tool for scanning localhost websites using a headless browser.

## Features

- **Headless Browser:** Real Chrome/Chromium execution for JS errors.
- **Concurrent Scanning:** Uses a worker pool to scan multiple internal pages in parallel.
- **Error Sniffing:** Catches JS Exceptions, Console Errors, and Network Failures.
- **Link Discovery:** Automatically crawls internal links on your domain.
- **CI/CD Friendly:** Returns non-zero exit codes on scan failure.

## Installation

### From Source

Ensure you have Go installed (1.20+), then:

```bash
git clone https://github.com/Woeter69/wcheck
cd wcheck
make install
```

This will build the binary and move it to `/usr/local/bin`.

## Usage

### Simple Scan

```bash
wcheck scan http://localhost:3000
```

### Advanced Scan

```bash
wcheck scan http://localhost:3000 --workers 10 --verbose
```

### Flags

- `-w, --workers [int]`: Set the number of concurrent worker goroutines (default: 5).
- `-v, --verbose`: Enable debug logs and see per-worker activity.

## Automation

### Git Hooks

Keep your production site safe with a pre-push hook. 

1. Copy the `pre-push.sh` script to your `.git/hooks/` directory:

   ```bash
   cp pre-push.sh .git/hooks/pre-push
   chmod +x .git/hooks/pre-push
   ```

2. Every time you run `git push`, `wcheck` will scan your local dev server and abort the push if any errors are found.

## License

MIT
