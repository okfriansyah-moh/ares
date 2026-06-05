#!/usr/bin/env bash
# install.sh - ARES one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh | bash
# Override install dir:   ARS_INSTALL_DIR=/usr/local/bin bash install.sh
# Pin a version:          ARS_VERSION=v1.2.0 bash install.sh

set -euo pipefail

REPO="okfriansyah-moh/ares"
INSTALL_DIR="${ARS_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="ars"
VERSION="${ARS_VERSION:-latest}"

if [ -t 1 ]; then
  GREEN='\033[32m'; BLUE='\033[34m'; YELLOW='\033[33m'; RED='\033[31m'; RESET='\033[0m'; BOLD='\033[1m'
else
  GREEN=''; BLUE=''; YELLOW=''; RED=''; RESET=''; BOLD=''
fi

info()  { printf '%b %s\n' "${BLUE}->${RESET}" "$*"; }
ok()    { printf '%b %s\n' "${GREEN}OK${RESET}" "$*"; }
warn()  { printf '%b %s\n' "${YELLOW}!${RESET}" "$*" >&2; }
fatal() { printf '%b %s\n' "${RED}Error:${RESET}" "$*" >&2; exit 1; }

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) fatal "unsupported architecture: ${ARCH}" ;;
esac

case "${OS}" in
  linux | darwin) ;;
  msys* | cygwin* | mingw*)
    fatal "Windows detected. Download the binary from: https://github.com/${REPO}/releases"
    ;;
  *) fatal "unsupported OS: ${OS}. Download from: https://github.com/${REPO}/releases" ;;
esac

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"

if [ "${VERSION}" = "latest" ]; then
  BASE_URL="https://github.com/${REPO}/releases/latest/download"
else
  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
fi

DOWNLOAD_URL="${BASE_URL}/${ASSET_NAME}"
CHECKSUM_URL="${BASE_URL}/${ASSET_NAME}.sha256"

info "Downloading ars (${OS}/${ARCH})..."
mkdir -p "${INSTALL_DIR}"

TMP_DIR="$(mktemp -d)"
TMP_BIN="${TMP_DIR}/${ASSET_NAME}"
TMP_SUM="${TMP_DIR}/${ASSET_NAME}.sha256"
trap 'rm -rf "${TMP_DIR}"' EXIT

curl -fsSL --progress-bar "${DOWNLOAD_URL}" -o "${TMP_BIN}" \
  || fatal "download failed: ${DOWNLOAD_URL}"

curl -fsSL "${CHECKSUM_URL}" -o "${TMP_SUM}" 2>/dev/null || true

if [ -s "${TMP_SUM}" ]; then
  info "Verifying checksum..."
  EXPECTED="$(awk '{print $1}' "${TMP_SUM}")"
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL="$(sha256sum "${TMP_BIN}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL="$(shasum -a 256 "${TMP_BIN}" | awk '{print $1}')"
  else
    warn "no sha256 tool found; skipping checksum verification"
    ACTUAL="${EXPECTED}"
  fi
  [ "${ACTUAL}" = "${EXPECTED}" ] \
    || fatal "checksum mismatch\n  expected: ${EXPECTED}\n  got:      ${ACTUAL}"
  ok "Checksum verified"
fi

chmod +x "${TMP_BIN}"
mv "${TMP_BIN}" "${INSTALL_DIR}/${BINARY_NAME}"
ok "Installed ars to ${INSTALL_DIR}/${BINARY_NAME}"

if command -v ars >/dev/null 2>&1; then
  printf '\n'
  ok "ars is already on your PATH."
  printf '\n  Verify: %sars --version%s\n' "${BOLD}" "${RESET}"
else
  printf '\n'
  case "${SHELL:-bash}" in
    */zsh)
      PATH_CMD="echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
      ;;
    */fish)
      PATH_CMD="fish_add_path \$HOME/.local/bin"
      ;;
    *)
      PATH_CMD="echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
      ;;
  esac
  printf '%b Add ars to your PATH:\n\n' "${YELLOW}->${RESET}"
  printf "  %s\n\n" "${PATH_CMD}"
  printf "  Then verify: %sars --version%s\n" "${BOLD}" "${RESET}"
fi
