#!/bin/sh
# pixelize installer.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/chris-roerig/pixelize/main/install.sh | sh
#
# Environment overrides:
#   PIXELIZE_VERSION  install a specific tag (e.g. v0.2.0); defaults to latest
#   PIXELIZE_BINDIR   install directory (default: /usr/local/bin, or ~/.local/bin
#                     if the former is not writable)
set -eu

REPO="chris-roerig/pixelize"
BINARY="pixelize"

info() { printf '%s\n' "$*" >&2; }
err() { printf 'error: %s\n' "$*" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"; }

# Pick a downloader.
download() {
	# download <url> <output>
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$2" "$1"
	else
		err "need curl or wget to download"
	fi
}

download_stdout() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO- "$1"
	else
		err "need curl or wget to download"
	fi
}

need uname
need tar

# Normalize OS to match the GoReleaser archive naming (title-cased uname -s).
os_raw="$(uname -s)"
case "$os_raw" in
	Linux) OS="Linux" ;;
	Darwin) OS="Darwin" ;;
	*) err "unsupported OS: $os_raw (install from source: https://github.com/$REPO)" ;;
esac

# Normalize arch to match the archive naming.
arch_raw="$(uname -m)"
case "$arch_raw" in
	x86_64 | amd64) ARCH="x86_64" ;;
	arm64 | aarch64) ARCH="arm64" ;;
	*) err "unsupported architecture: $arch_raw" ;;
esac

# Resolve version.
VERSION="${PIXELIZE_VERSION:-}"
if [ -z "$VERSION" ]; then
	info "Resolving latest release..."
	# Parse the tag_name from the GitHub API without requiring jq.
	VERSION="$(download_stdout "https://api.github.com/repos/$REPO/releases/latest" \
		| grep -m1 '"tag_name"' \
		| sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
	[ -n "$VERSION" ] || err "could not determine latest version (set PIXELIZE_VERSION)"
fi

ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/$REPO/releases/download/$VERSION"
ARCHIVE_URL="$BASE_URL/$ARCHIVE"
CHECKSUMS_URL="$BASE_URL/checksums.txt"

info "Installing $BINARY $VERSION ($OS/$ARCH)"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT INT TERM

download "$ARCHIVE_URL" "$TMP/$ARCHIVE" || err "download failed: $ARCHIVE_URL"

# Verify checksum when a sha256 tool and checksums file are available.
if download "$CHECKSUMS_URL" "$TMP/checksums.txt" 2>/dev/null; then
	sum=""
	if command -v sha256sum >/dev/null 2>&1; then
		sum="$(sha256sum "$TMP/$ARCHIVE" | awk '{print $1}')"
	elif command -v shasum >/dev/null 2>&1; then
		sum="$(shasum -a 256 "$TMP/$ARCHIVE" | awk '{print $1}')"
	fi
	if [ -n "$sum" ]; then
		want="$(grep " $ARCHIVE\$" "$TMP/checksums.txt" | awk '{print $1}')"
		if [ -n "$want" ] && [ "$sum" != "$want" ]; then
			err "checksum mismatch for $ARCHIVE"
		fi
		[ -n "$want" ] && info "Checksum verified."
	fi
else
	info "Warning: checksums.txt not found; skipping verification."
fi

tar -xzf "$TMP/$ARCHIVE" -C "$TMP" || err "failed to extract archive"
[ -f "$TMP/$BINARY" ] || err "binary not found in archive"
chmod +x "$TMP/$BINARY"

# Choose an install directory.
BINDIR="${PIXELIZE_BINDIR:-}"
if [ -z "$BINDIR" ]; then
	if [ -w /usr/local/bin ] 2>/dev/null; then
		BINDIR="/usr/local/bin"
	elif [ "$(id -u)" = "0" ]; then
		BINDIR="/usr/local/bin"
	else
		BINDIR="$HOME/.local/bin"
	fi
fi
mkdir -p "$BINDIR"

if mv "$TMP/$BINARY" "$BINDIR/$BINARY" 2>/dev/null; then
	:
elif command -v sudo >/dev/null 2>&1 && [ "$BINDIR" = "/usr/local/bin" ]; then
	info "Elevating with sudo to write to $BINDIR..."
	sudo mv "$TMP/$BINARY" "$BINDIR/$BINARY"
else
	err "cannot write to $BINDIR (set PIXELIZE_BINDIR to a writable path)"
fi

info "Installed $BINARY to $BINDIR/$BINARY"

# Install shell completions if the archive bundled them (best-effort, user-local).
if [ -d "$TMP/completions" ]; then
	# zsh: a dir on the default fpath for user functions.
	zsh_dir="$HOME/.zsh/completions"
	if [ -f "$TMP/completions/_pixelize" ]; then
		mkdir -p "$zsh_dir" 2>/dev/null && cp "$TMP/completions/_pixelize" "$zsh_dir/_pixelize" 2>/dev/null && \
			info "Installed zsh completion to $zsh_dir/_pixelize"
	fi
	# bash: user bash-completion dir.
	bash_dir="${BASH_COMPLETION_USER_DIR:-$HOME/.local/share/bash-completion}/completions"
	if [ -f "$TMP/completions/pixelize.bash" ]; then
		mkdir -p "$bash_dir" 2>/dev/null && cp "$TMP/completions/pixelize.bash" "$bash_dir/pixelize" 2>/dev/null && \
			info "Installed bash completion to $bash_dir/pixelize"
	fi
	# fish: user completions dir.
	fish_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions"
	if [ -f "$TMP/completions/pixelize.fish" ]; then
		mkdir -p "$fish_dir" 2>/dev/null && cp "$TMP/completions/pixelize.fish" "$fish_dir/pixelize.fish" 2>/dev/null && \
			info "Installed fish completion to $fish_dir/pixelize.fish"
	fi
	info "Restart your shell (or source the completion) to enable tab completion."
	info "zsh users: ensure 'fpath=(~/.zsh/completions \$fpath)' is set before compinit in ~/.zshrc."
fi

case ":$PATH:" in
	*":$BINDIR:"*) ;;
	*) info "Note: $BINDIR is not on your PATH. Add it, e.g.:"
	   info "  export PATH=\"$BINDIR:\$PATH\"" ;;
esac

"$BINDIR/$BINARY" --version 2>/dev/null || true
