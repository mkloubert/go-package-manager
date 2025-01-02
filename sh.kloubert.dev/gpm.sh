#!/bin/bash

handle_error() {
    echo "Error: $1"
    exit 1
}

echo "Go Package Manager Updater"
echo ""

GPM_BIN_PATH=${GPM_BIN_PATH:-/usr/local/bin}

case "$(uname -s)" in
    Darwin)
        OS="darwin"
        ;;
    Linux)
        OS="linux"
        ;;
    FreeBSD)
        OS="freebsd"
        ;;
    OpenBSD)
        OS="openbsd"
        ;;
    CYGWIN*|MINGW*|MSYS*|Windows_NT)
        OS="windows"
        ;;
    *)
        OS="unknown"
        echo "Error: Unknown operating system detected!"
        exit 1
        ;;
esac

echo "Your operating system: $OS"

# Detect the system architecture
case "$(uname -m)" in
    x86_64)
        ARCH="amd64"
        ;;
    i*86)
        ARCH="386"
        ;;
    armv7*|armv6*|armv5*)
        ARCH="arm"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        ARCH="unknown"
        echo "Error: Unknown architecture detected!"
        exit 1
        ;;
esac

echo "Your architecture: $ARCH"
echo ""

echo "Finding download URL and SHA256 URL ..."
latest_release_info=$(wget -qO- https://api.github.com/repos/mkloubert/go-package-manager/releases/latest) || handle_error "Could not fetch release infos"

download_url=$(echo "$latest_release_info" | jq -r \
    ".assets[].browser_download_url | select(contains(\"gpm\") and contains(\"$OS\") and contains(\"$ARCH\") and (. | tostring | contains(\"sha256\") | not))") \
    || handle_error "Could not parse download URL"
sha256_url=$(echo "$latest_release_info" | jq -r \
    ".assets[].browser_download_url | select(contains(\"gpm\") and contains(\"$OS\") and contains(\"$ARCH\") and contains(\"sha256\"))") \
    || handle_error "Could not parse SHA256 URL"

if [ -z "$download_url" ]; then
  handle_error "No valid download URL found"
fi

if [ -z "$sha256_url" ]; then
  handle_error "No valid SHA256 URL found"
fi

echo "Downloading tarball from '$download_url'..."
wget -q "$download_url" -O gpm.tar.gz || handle_error "Failed to download tarball"

echo "Downloading SHA256 file from '$sha256_url'..."
wget -q "$sha256_url" -O gpm.tar.gz.sha256 || handle_error "Failed to download SHA256 file"

echo "Verifying tarball ..."
shasum -a 256 gpm.tar.gz.sha256 || handle_error "SHA256 verification failed"

echo "Extracting binary ..."
tar -xzOf gpm.tar.gz gpm > gpm || handle_error "Could not extract 'gpm' binary"

echo "Installing 'gpm' to $GPM_BIN_PATH ..."
sudo mv gpm "$GPM_BIN_PATH/gpm" || handle_error "Could not move 'gpm' to '$GPM_BIN_PATH'"
sudo chmod +x "$GPM_BIN_PATH/gpm" || handle_error "Could not update permissions of 'gpm' binary"

echo "Cleaning up ..."
rm gpm.tar.gz gpm.tar.gz.sha256 || handle_error "Cleanups failed"

echo "'gpm' successfully installed or updated 👍"
