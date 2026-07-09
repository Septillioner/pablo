#!/usr/bin/env bash
set -euo pipefail

# Pablo one-liner installer (macOS / Linux)
# Usage: curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/main/install.sh | bash
# Pin version: PABLO_VERSION=v1.4.0 curl -fsSL ... | bash

GITHUB_REPO="septillioner/pablo"
GITHUB_API_BASE="https://api.github.com/repos/${GITHUB_REPO}/releases"
RELEASES_PAGE="https://github.com/${GITHUB_REPO}/releases"

SYSTEM_INSTALL_DIR="/usr/local/bin"
SYSTEM_INSTALL_PATH="${SYSTEM_INSTALL_DIR}/pablo"
USER_INSTALL_DIR="${HOME}/.local/bin"
USER_INSTALL_PATH="${USER_INSTALL_DIR}/pablo"

log() {
  printf '==> %s\n' "$1"
}

fail() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "required command not found: $1"
  fi
}

detect_platform() {
  local os arch

  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "${os}" in
    darwin) os="darwin" ;;
    linux) os="linux" ;;
    *) fail "unsupported operating system: ${os}. See ${RELEASES_PAGE}" ;;
  esac

  case "${arch}" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *) fail "unsupported architecture: ${arch}. See ${RELEASES_PAGE}" ;;
  esac

  ASSET_NAME="pablo-${os}-${arch}"
  PLATFORM_LABEL="${os}/${arch}"
}

resolve_release_tag() {
  if [[ -n "${PABLO_VERSION:-}" ]]; then
    RELEASE_TAG="${PABLO_VERSION}"
    [[ "${RELEASE_TAG}" == v* ]] || RELEASE_TAG="v${RELEASE_TAG}"
    return
  fi

  require_command curl
  RELEASE_TAG="$(
    curl -fsSL "${GITHUB_API_BASE}/latest" \
      | awk -F'"' '/"tag_name":/ { print $4; exit }'
  )"
  [[ -n "${RELEASE_TAG}" ]] || fail "could not resolve latest release tag"
}

parse_asset_url() {
  local release_json="$1"
  local asset_name="$2"

  printf '%s' "${release_json}" | awk -v asset="${asset_name}" '
    /"name":/ {
      line = $0
      sub(/.*"name":[[:space:]]*"/, "", line)
      sub(/".*/, "", line)
      current = line
    }
    /"browser_download_url":/ && current == asset {
      line = $0
      sub(/.*"browser_download_url":[[:space:]]*"/, "", line)
      sub(/".*/, "", line)
      print line
      exit
    }
  '
}

fetch_asset_url() {
  local release_json asset_url checksum_url

  require_command curl
  release_json="$(curl -fsSL "${GITHUB_API_BASE}/tags/${RELEASE_TAG}")"

  asset_url="$(parse_asset_url "${release_json}" "${ASSET_NAME}")"
  checksum_url="$(parse_asset_url "${release_json}" "checksums.txt")"

  [[ -n "${asset_url}" ]] || fail "release ${RELEASE_TAG} has no asset ${ASSET_NAME}. See ${RELEASES_PAGE}"
  ASSET_URL="${asset_url}"
  CHECKSUMS_URL="${checksum_url}"
}

download_binary() {
  local tmp_dir

  require_command curl
  require_command mktemp

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "${tmp_dir}"' EXIT

  log "downloading ${ASSET_NAME} from ${RELEASE_TAG}"
  curl -fsSL "${ASSET_URL}" -o "${tmp_dir}/${ASSET_NAME}"
  chmod +x "${tmp_dir}/${ASSET_NAME}"

  if [[ -n "${CHECKSUMS_URL:-}" ]]; then
    verify_checksum "${tmp_dir}"
  else
    log "checksums.txt not found in release; skipping verification"
  fi

  DOWNLOADED_BINARY="${tmp_dir}/${ASSET_NAME}"
}

verify_checksum() {
  local tmp_dir checksum_file expected_hash actual_hash

  tmp_dir="$(dirname "${DOWNLOADED_BINARY}")"
  checksum_file="${tmp_dir}/checksums.txt"

  curl -fsSL "${CHECKSUMS_URL}" -o "${checksum_file}"

  expected_hash="$(
    awk -v asset="${ASSET_NAME}" '$2 == asset { print $1; exit }' "${checksum_file}"
  )"
  [[ -n "${expected_hash}" ]] || fail "checksum for ${ASSET_NAME} not found in checksums.txt"

  if command -v sha256sum >/dev/null 2>&1; then
    actual_hash="$(sha256sum "${DOWNLOADED_BINARY}" | awk '{ print $1 }')"
  elif command -v shasum >/dev/null 2>&1; then
    actual_hash="$(shasum -a 256 "${DOWNLOADED_BINARY}" | awk '{ print $1 }')"
  else
    fail "sha256sum or shasum is required for checksum verification"
  fi

  if [[ "${actual_hash}" != "${expected_hash}" ]]; then
    fail "checksum mismatch for ${ASSET_NAME}"
  fi

  log "checksum verified"
}

install_to_system() {
  if [[ -w "${SYSTEM_INSTALL_DIR}" ]]; then
    install -m 0755 "${DOWNLOADED_BINARY}" "${SYSTEM_INSTALL_PATH}"
    INSTALLED_PATH="${SYSTEM_INSTALL_PATH}"
    INSTALL_SCOPE="system"
    return 0
  fi

  if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    sudo install -m 0755 "${DOWNLOADED_BINARY}" "${SYSTEM_INSTALL_PATH}"
    INSTALLED_PATH="${SYSTEM_INSTALL_PATH}"
    INSTALL_SCOPE="system"
    return 0
  fi

  return 1
}

install_to_user() {
  mkdir -p "${USER_INSTALL_DIR}"
  install -m 0755 "${DOWNLOADED_BINARY}" "${USER_INSTALL_PATH}"
  INSTALLED_PATH="${USER_INSTALL_PATH}"
  INSTALL_SCOPE="user"
}

path_contains_dir() {
  local dir="$1"
  case ":${PATH}:" in
    *":${dir}:"*) return 0 ;;
    *) return 1 ;;
  esac
}

ensure_user_path() {
  local profile_file marker_line shell_name

  if path_contains_dir "${USER_INSTALL_DIR}"; then
    return 0
  fi

  marker_line="export PATH=\"\${HOME}/.local/bin:\${PATH}\""
  shell_name="$(basename "${SHELL:-}")"

  case "${shell_name}" in
    zsh) profile_file="${HOME}/.zshrc" ;;
    bash) profile_file="${HOME}/.bashrc" ;;
    *) profile_file="${HOME}/.profile" ;;
  esac

  if [[ -f "${profile_file}" ]] && grep -Fq "${marker_line}" "${profile_file}"; then
    return 0
  fi

  printf '\n# Added by Pablo installer\n%s\n' "${marker_line}" >> "${profile_file}"
  log "added ${USER_INSTALL_DIR} to PATH in ${profile_file}"
  log "open a new shell or run: export PATH=\"${USER_INSTALL_DIR}:\$PATH\""
}

verify_installation() {
  if command -v pablo >/dev/null 2>&1; then
    pablo version
    return 0
  fi

  if [[ -x "${INSTALLED_PATH}" ]]; then
    "${INSTALLED_PATH}" version
    if [[ "${INSTALL_SCOPE}" == "user" ]]; then
      log "pablo is installed at ${INSTALLED_PATH}; reload your shell to use it globally"
    fi
    return 0
  fi

  fail "installation finished but pablo could not be executed"
}

main() {
  detect_platform
  resolve_release_tag
  fetch_asset_url
  download_binary

  log "installing Pablo for ${PLATFORM_LABEL} (${RELEASE_TAG})"

  if install_to_system; then
    log "installed to ${INSTALLED_PATH} (system)"
  else
    log "system install unavailable; using user install"
    install_to_user
    ensure_user_path
    log "installed to ${INSTALLED_PATH} (user)"
  fi

  verify_installation
  log "done"
}

main "$@"
