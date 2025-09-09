#!/usr/bin/env bash

set -euo pipefail

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $OS in
    linux)
        case $ARCH in
            x86_64) BINARY="habits-linux" ;;
            aarch64|arm64) BINARY="habits-linux-arm64" ;;
            *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    darwin) BINARY="habits-macos" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "Downloading $BINARY..."
curl -SsL "https://github.com/brk3/habits/releases/latest/download/$BINARY" -o habits
chmod +x habits

echo "Installing to /usr/local/bin..."
sudo mv habits /usr/local/bin/
