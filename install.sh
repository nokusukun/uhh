#!/bin/bash
# UHH Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/nokusukun/uhh/main/install.sh | bash

set -e

REPO="nokusukun/uhh"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    linux)
        PLATFORM="linux"
        ;;
    darwin)
        PLATFORM="darwin"
        ;;
    *)
        echo "Unsupported OS: $OS"
        echo "Please download manually from https://github.com/$REPO/releases"
        exit 1
        ;;
esac

echo "Detected: $PLATFORM-$ARCH"

# Get latest release version
echo "Fetching latest release..."
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo "Failed to fetch latest version. Please install manually."
    exit 1
fi

echo "Latest version: $LATEST_VERSION"

# Download
FILENAME="uhh-${LATEST_VERSION}-${PLATFORM}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$FILENAME"

echo "Downloading $DOWNLOAD_URL..."
TEMP_DIR=$(mktemp -d)
curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/$FILENAME"

# Extract
echo "Extracting..."
tar -xzf "$TEMP_DIR/$FILENAME" -C "$TEMP_DIR"

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_DIR/uhh" "$INSTALL_DIR/uhh"
else
    echo "Need sudo to install to $INSTALL_DIR"
    sudo mv "$TEMP_DIR/uhh" "$INSTALL_DIR/uhh"
fi

chmod +x "$INSTALL_DIR/uhh"

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "UHH $LATEST_VERSION installed successfully!"
echo ""
echo "Run 'uhh init' to configure your LLM providers."
echo "Run 'uhh --help' for usage information."
