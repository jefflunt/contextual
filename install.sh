#!/bin/bash
set -e

REPO="jefflunt/contextual"

echo "Detecting OS and Architecture..."
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)
        OS_NAME="linux"
        ;;
    Darwin)
        OS_NAME="darwin"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64)
        ARCH_NAME="amd64"
        ;;
    arm64|aarch64)
        ARCH_NAME="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

BINARY_NAME="contextual-${OS_NAME}-${ARCH_NAME}"

echo "Finding latest release for ${OS_NAME}/${ARCH_NAME}..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not find the latest release."
    exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY_NAME}"

# Try to install to /usr/local/bin (requires sudo), fallback to ~/.local/bin
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    echo "Warning: /usr/local/bin is not writable. Installing to $INSTALL_DIR instead."
    echo "Make sure $INSTALL_DIR is in your PATH."
fi

echo "Downloading ${BINARY_NAME}..."
curl -sL "$DOWNLOAD_URL" -o "contextual"
chmod +x contextual

echo "Installing to ${INSTALL_DIR}..."
if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
    sudo mv contextual "$INSTALL_DIR/contextual"
else
    mv contextual "$INSTALL_DIR/contextual"
fi

echo "Successfully installed contextual to ${INSTALL_DIR}/contextual"
echo "Run 'contextual version' to verify."
