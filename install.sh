#!/bin/sh
# Installs the latest ccw release binary from GitHub.
#
#   curl -fsSL https://raw.githubusercontent.com/glstep/claude-with/main/install.sh | sh
#
# Environment:
#   BINDIR  install directory (default: /usr/local/bin)

set -eu

repo="glstep/claude-with"
bindir="${BINDIR:-/usr/local/bin}"

if ! command -v curl >/dev/null 2>&1; then
    echo "error: curl is required" >&2
    exit 1
fi

os=$(uname -s)
case "$os" in
Linux) os=linux ;;
Darwin) os=darwin ;;
*)
    echo "error: unsupported OS: $os" >&2
    echo "On Windows, download ccw_windows_amd64.exe from https://github.com/$repo/releases/latest" >&2
    exit 1
    ;;
esac

arch=$(uname -m)
case "$arch" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*)
    echo "error: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

url="https://github.com/$repo/releases/latest/download/ccw_${os}_${arch}"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

echo "Downloading $url"
if ! curl -fsSL -o "$tmpdir/ccw" "$url"; then
    echo "error: download failed — is there a release with a ${os}/${arch} build?" >&2
    echo "See https://github.com/$repo/releases/latest for available assets." >&2
    exit 1
fi
chmod +x "$tmpdir/ccw"

if [ ! -d "$bindir" ]; then
    mkdir -p "$bindir" 2>/dev/null || true
fi

if [ -d "$bindir" ] && [ -w "$bindir" ]; then
    mv "$tmpdir/ccw" "$bindir/ccw"
elif command -v sudo >/dev/null 2>&1; then
    echo "Installing to $bindir (requires sudo)"
    sudo mkdir -p "$bindir"
    sudo mv "$tmpdir/ccw" "$bindir/ccw"
else
    echo "error: $bindir is not writable and sudo is unavailable." >&2
    echo "Re-run with BINDIR set to a writable directory, e.g.:" >&2
    echo "  curl -fsSL https://raw.githubusercontent.com/$repo/main/install.sh | BINDIR=\$HOME/.local/bin sh" >&2
    exit 1
fi

echo "Installed $("$bindir/ccw" --version) to $bindir/ccw"
