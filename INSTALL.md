# Installation Guide

`wcheck` is built with Go and requires a Chromium-based browser (Chrome, Chromium, or Brave) installed on your system to operate.

## Prerequisites

- **Go:** Version 1.23 or higher.
- **Chrome/Chromium:** Required for headless execution.

### Arch Linux (Pacman)

On Arch Linux, you can install the necessary requirements using `pacman`:

```bash
sudo pacman -S chromium go make
```

### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install -y google-chrome-stable golang-go make
```

## Install from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/Woeter69/wcheck.git
   cd wcheck
   ```

2. Build the binary:
   ```bash
   make build
   ```

3. Install to your system path:
   ```bash
   sudo make install
   ```

## Verify Installation

Check if the tool is installed and accessible:

```bash
wcheck --version
wcheck --help
```
