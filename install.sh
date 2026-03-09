#!/bin/sh
set -e

REPO="HenriqueSchroeder/beacon"
INSTALL_DIR="/usr/local/bin"
BINARY="beacon"

main() {
    os="$(detect_os)"
    arch="$(detect_arch)"
    version="$(latest_version)"

    if [ -z "$version" ]; then
        echo "Error: could not determine latest version." >&2
        exit 1
    fi

    archive="${BINARY}_${os}_${arch}.tar.gz"
    if [ "$os" = "windows" ]; then
        archive="${BINARY}_${os}_${arch}.zip"
    fi

    url="https://github.com/${REPO}/releases/download/${version}/${archive}"

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    echo "Downloading beacon ${version} for ${os}/${arch}..."
    download "$url" "$tmpdir/$archive"

    echo "Extracting..."
    if [ "$os" = "windows" ]; then
        unzip -q "$tmpdir/$archive" -d "$tmpdir"
    else
        tar xzf "$tmpdir/$archive" -C "$tmpdir"
    fi

    install_binary "$tmpdir/$BINARY"

    echo "beacon ${version} installed successfully."
    echo "Run 'beacon version' to verify."
}

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) echo "Error: unsupported OS '$(uname -s)'" >&2; exit 1 ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "Error: unsupported architecture '$(uname -m)'" >&2; exit 1 ;;
    esac
}

latest_version() {
    if command -v curl > /dev/null 2>&1; then
        curl -sI "https://github.com/${REPO}/releases/latest" | grep -i "^location:" | sed 's|.*/||' | tr -d '\r'
    elif command -v wget > /dev/null 2>&1; then
        wget --spider -S "https://github.com/${REPO}/releases/latest" 2>&1 | grep -i "^  location:" | sed 's|.*/||' | tr -d '\r'
    else
        echo "Error: curl or wget required." >&2
        exit 1
    fi
}

download() {
    if command -v curl > /dev/null 2>&1; then
        curl -sL "$1" -o "$2"
    elif command -v wget > /dev/null 2>&1; then
        wget -q "$1" -O "$2"
    fi
}

install_binary() {
    if [ -w "$INSTALL_DIR" ]; then
        mv "$1" "$INSTALL_DIR/$BINARY"
        chmod +x "$INSTALL_DIR/$BINARY"
    elif command -v sudo > /dev/null 2>&1; then
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$1" "$INSTALL_DIR/$BINARY"
        sudo chmod +x "$INSTALL_DIR/$BINARY"
    else
        fallback_dir="$HOME/.local/bin"
        mkdir -p "$fallback_dir"
        mv "$1" "$fallback_dir/$BINARY"
        chmod +x "$fallback_dir/$BINARY"
        echo "Installed to $fallback_dir (add to PATH if needed)."
    fi
}

main
