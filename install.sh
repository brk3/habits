#!/usr/bin/env bash

set -euo pipefail

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
DEST="/usr/local/bin"

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

# make a new tempfile and the accompanying cleanup trap
tmpfile="$(mktemp)"
trap 'rm -rf "${tmpfile}"' EXIT

echo "Downloading $BINARY..."
curl -SsL "https://github.com/brk3/habits/releases/latest/download/$BINARY" -o "${tmpfile}"

echo "Installing to ${DEST}..."
sudo install -m 0755 -o root -g root "${tmpfile}" "${DEST}/habits"
