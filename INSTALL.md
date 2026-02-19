# Installation Guide

`wcheck` is built with Go and requires a Chromium-based browser to operate.

## Dependencies
Ensure you have `chromium` or `google-chrome` installed on your system.

### Arch Linux
```bash
sudo pacman -S chromium
```

## Build from Source
1. **Clone the repository:**
   ```bash
   git clone https://github.com/Woeter69/wcheck.git
   cd wcheck
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Install:**
   The included `Makefile` handles the build and system installation.
   ```bash
   make install
   ```
   *Note: This will build the binary and move it to `/usr/local/bin` using `sudo`.*

## Verification
Verify the installation by checking the version or help menu:
```bash
wcheck --help
```
