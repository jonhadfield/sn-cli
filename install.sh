#!/bin/sh
# Install sn-cli.
#
#   curl -fsSL https://raw.githubusercontent.com/jonhadfield/sn-cli/main/install.sh | sh
#
# Installs the latest release into /usr/local/bin, using sudo only for that
# final step. Override the destination with BIN_DIR, or pin a release with
# VERSION:
#
#   curl -fsSL .../install.sh | BIN_DIR="$HOME/.local/bin" sh
#   curl -fsSL .../install.sh | VERSION=0.5.0 sh

set -eu

REPO="jonhadfield/sn-cli"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"

fail() {
    echo "install.sh: $1" >&2
    exit 1
}

# Work out which published archive matches this machine.
os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
    Darwin) asset_os="Darwin" ;;
    Linux) asset_os="Linux" ;;
    FreeBSD) asset_os="Freebsd" ;;
    OpenBSD) asset_os="Openbsd" ;;
    *) fail "unsupported operating system: $os" ;;
esac

if [ "$asset_os" = "Darwin" ]; then
    # There is one universal darwin build rather than a per-architecture one.
    asset="sn-cli_Darwin_universal.tar.gz"
else
    case "$arch" in
        x86_64 | amd64) asset_arch="x86_64" ;;
        aarch64 | arm64) asset_arch="arm64" ;;
        i386 | i686) asset_arch="i386" ;;
        *) fail "unsupported architecture: $arch" ;;
    esac

    asset="sn-cli_${asset_os}_${asset_arch}.tar.gz"
fi

# Resolve the latest tag from the release redirect rather than the API, which
# is rate limited for unauthenticated callers.
version="${VERSION:-}"
if [ -z "$version" ]; then
    latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")" ||
        fail "could not reach github to determine the latest release"
    version="${latest_url##*/tag/}"
fi

case "$version" in
    "" | */*) fail "could not determine the release version" ;;
esac

base="https://github.com/$REPO/releases/download/$version"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "Downloading sn-cli $version ($asset)"
curl -fsSL "$base/$asset" -o "$tmp/$asset" || fail "failed to download $base/$asset"

# Verify against the published checksums before running anything.
if curl -fsSL "$base/sn-cli_${version}_checksums.txt" -o "$tmp/checksums.txt"; then
    expected="$(awk -v a="$asset" '$2 == a || $2 == "*"a {print $1}' "$tmp/checksums.txt")"

    if [ -z "$expected" ]; then
        fail "no checksum published for $asset"
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        actual="$(sha256sum "$tmp/$asset" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual="$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')"
    else
        fail "need sha256sum or shasum to verify the download"
    fi

    if [ "$expected" != "$actual" ]; then
        fail "checksum mismatch for $asset: expected $expected, got $actual"
    fi

    echo "Checksum verified"
else
    fail "could not download checksums for $version"
fi

tar -xzf "$tmp/$asset" -C "$tmp" sn || fail "archive did not contain sn"
chmod 755 "$tmp/sn"

# macOS marks downloads as quarantined and these binaries are unsigned, so
# Gatekeeper would block the first run.
if [ "$asset_os" = "Darwin" ] && command -v xattr >/dev/null 2>&1; then
    xattr -d com.apple.quarantine "$tmp/sn" 2>/dev/null || true
fi

sudo=""
if [ ! -d "$BIN_DIR" ] || [ ! -w "$BIN_DIR" ]; then
    if [ "$(id -u)" -ne 0 ]; then
        command -v sudo >/dev/null 2>&1 || fail "$BIN_DIR is not writable and sudo is not available"
        sudo="sudo"
    fi
fi

$sudo mkdir -p "$BIN_DIR"
$sudo install -m 755 "$tmp/sn" "$BIN_DIR/sn" || fail "failed to install into $BIN_DIR"

echo "Installed sn to $BIN_DIR/sn"
"$BIN_DIR/sn" --version
