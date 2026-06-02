#!/usr/bin/env bash
set -euo pipefail

REPO="adcreafy/skill-chrome"
BINARY_NAME="skill-chrome-host"
HOST_NAME="com.nicepkg.skill_chrome"
INSTALL_DIR="${HOME}/.skill-chrome"

# Detect OS and Architecture
detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $os" && exit 1 ;;
  esac

  case "$arch" in
    x86_64|amd64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $arch" && exit 1 ;;
  esac
}

get_latest_version() {
  VERSION=""
  if command -v curl >/dev/null 2>&1; then
    local json
    json=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null) || true
    if [ -n "$json" ]; then
      VERSION=$(echo "$json" | grep '"tag_name"' | head -1 | cut -d'"' -f4) || true
    fi
  fi
  if [ -z "$VERSION" ]; then
    VERSION="v0.1.0"
  fi
}

download_binary() {
  local suffix="${OS}-${ARCH}"
  [ "$OS" = "windows" ] && suffix="${suffix}.exe"
  local url="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${suffix}"

  echo "Downloading ${BINARY_NAME} ${VERSION} for ${OS}/${ARCH}..."
  mkdir -p "$INSTALL_DIR"

  local target="${INSTALL_DIR}/${BINARY_NAME}"
  [ "$OS" = "windows" ] && target="${target}.exe"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$target"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$target"
  else
    echo "Error: curl or wget is required"
    exit 1
  fi

  chmod +x "$target"
  echo "Binary installed to: $target"
}

register_native_manifest() {
  local binary_path="${INSTALL_DIR}/${BINARY_NAME}"

  local manifest='{
  "name": "'"${HOST_NAME}"'",
  "description": "Skill Chrome Native Messaging Host",
  "path": "'"${binary_path}"'",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://lljmodipbnojpafcaegnlmfbanncbdpj/"
  ]
}'

  local manifest_dir manifest_path

  case "$OS" in
    darwin)
      # Chrome
      manifest_dir="${HOME}/Library/Application Support/Google/Chrome/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      manifest_path="${manifest_dir}/${HOST_NAME}.json"
      echo "$manifest" > "$manifest_path"
      echo "Chrome manifest: $manifest_path"

      # Chromium
      manifest_dir="${HOME}/Library/Application Support/Chromium/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      echo "$manifest" > "${manifest_dir}/${HOST_NAME}.json"

      # Edge
      manifest_dir="${HOME}/Library/Application Support/Microsoft Edge/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      echo "$manifest" > "${manifest_dir}/${HOST_NAME}.json"

      # Arc (uses Chrome profile)
      manifest_dir="${HOME}/Library/Application Support/Arc/User Data/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      echo "$manifest" > "${manifest_dir}/${HOST_NAME}.json"
      ;;
    linux)
      manifest_dir="${HOME}/.config/google-chrome/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      manifest_path="${manifest_dir}/${HOST_NAME}.json"
      echo "$manifest" > "$manifest_path"
      echo "Chrome manifest: $manifest_path"

      manifest_dir="${HOME}/.config/chromium/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      echo "$manifest" > "${manifest_dir}/${HOST_NAME}.json"

      manifest_dir="${HOME}/.config/microsoft-edge/NativeMessagingHosts"
      mkdir -p "$manifest_dir"
      echo "$manifest" > "${manifest_dir}/${HOST_NAME}.json"
      ;;
    *)
      echo "Windows: please run install.ps1 instead"
      exit 1
      ;;
  esac
}

main() {
  echo "=== Skill Chrome Host Installer ==="
  echo ""

  detect_platform
  get_latest_version
  download_binary
  register_native_manifest

  echo ""
  echo "Installation complete!"
  echo "You can now use the Skill Chrome extension."
  echo ""
  echo "To uninstall, run:"
  echo "  rm -rf ${INSTALL_DIR}"
  echo "  rm -f ~/Library/Application\\ Support/Google/Chrome/NativeMessagingHosts/${HOST_NAME}.json"
}

main "$@"
