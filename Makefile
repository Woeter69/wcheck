# Makefile for wcheck

BINARY_NAME=wcheck
INSTALL_PATH=/usr/local/bin

.PHONY: all build install clean uninstall help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete. Run '$(BINARY_NAME) --help' to get started."

uninstall:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_PATH)..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

help:
	@echo "Usage:"
	@echo "  make build      - Compile the binary"
	@echo "  make install    - Build and install to $(INSTALL_PATH)"
	@echo "  make uninstall  - Remove the binary from $(INSTALL_PATH)"
	@echo "  make clean      - Remove build artifacts"
