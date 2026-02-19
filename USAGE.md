# Usage Guide

`wcheck` provides two main commands: `scan` for crawling/bulk auditing and `interact` for deep-diving into a single page.

## Command: `scan`
The `scan` command crawls the target URL for internal links and audits them concurrently.

### Flags
| Flag | Shorthand | Default | Description |
|------|-----------|---------|-------------|
| `--workers` | `-w` | `5` | Number of parallel browser instances. |
| `--timeout` | `-t` | `30` | Timeout in seconds for each page scan. |
| `--interact` | `-i` | `false` | Enable automatic "Monkey Testing" (clicking elements). |
| `--max-clicks`| `-m` | `20` | Max interactive elements to click per page. |
| `--delay` | `-d` | `0` | Delay in seconds between scans (useful for rate-limiting). |
| `--verbose` | `-v` | `false` | Enable detailed debug logging. |

### Example
```bash
wcheck scan http://localhost:3000 --workers 10 --interact
```

---

## Command: `interact`
Focuses on a single page and performs an exhaustive interaction test.

### Flags
- `-t, --timeout`: Default `60` (longer for deep interaction).
- `-m, --max-clicks`: Default `20`.
- `-v, --verbose`: View every interaction step.

```bash
wcheck interact http://localhost:3000/dashboard -v
```

---

## CI/CD Integration

### Exit Codes
`wcheck` is designed for automation:
- **0**: Scan complete with zero browser/network errors.
- **1**: Errors detected (JS exceptions, failed assets, or timeouts).

### Git Pre-Push Hook
Prevent broken code from reaching the server by adding `wcheck` to your hooks.
Create or edit `.git/hooks/pre-push`:

```bash
#!/bin/bash
echo "Running wcheck audit..."
wcheck scan http://localhost:3000 -i -w 10
if [ $? -ne 0 ]; then
 echo "Audit failed. Please fix the errors before pushing."
 exit 1
fi
```

## Deep Dive Reporting
The reporter categorizes errors into:
- **JS Error**: Uncaught exceptions in the browser.
- **Console**: `console.error` calls from the application.
- **Network**: Assets (JS, CSS, Images) that failed to load (404/500).
