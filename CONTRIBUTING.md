# Contributing to wcheck

Thank you for your interest in improving `wcheck`!

## Project Structure
- `main.go`: Entry point for the CLI.
- `cmd/`: Command definitions (Cobra).
  - `scan.go`: Bulk scanning and worker pool logic.
  - `interact.go`: Single-page interaction logic.
- `internal/`: Core logic (private to this module).
  - `engine/`: `chromedp` implementation and error sniffing.
  - `crawler/`: Internal link discovery logic.
  - `reporter/`: Terminal output and formatting.

## Development Workflow
1. **Prerequisites:** Go 1.21+ and Chromium.
2. **Setup:**
   ```bash
   go mod tidy
   ```
3. **Run during development:**
   ```bash
   go run main.go scan http://localhost:3000
   ```
4. **Testing:**
   Ensure your changes don't break the build:
   ```bash
   make build
   ```

## Guidelines
- Follow standard Go idioms and `gofmt`.
- Ensure all new features are accompanied by appropriate flags in `cmd/`.
- Maintain the minimalist, high-signal reporting style.
