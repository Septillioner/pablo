#!/usr/bin/env bash
set -euo pipefail

# Pablo one-liner installer (macOS / Linux)
# Usage: curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/scripts/install.sh | bash
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
  # Fetch fully before parsing. `curl | awk … exit` closes the pipe early →
  # curl (23) under `set -o pipefail`, which aborts the installer.
  local latest_json
  latest_json="$(curl -fsSL "${GITHUB_API_BASE}/latest")" \
    || fail "could not fetch latest release metadata"
  RELEASE_TAG="$(awk -F'"' '/"tag_name":/ { print $4; exit }' <<< "${latest_json}")"
  [[ -n "${RELEASE_TAG}" ]] || fail "could not resolve latest release tag"
}

parse_asset_url() {
  local release_json="$1"
  local asset_name="$2"

  awk -v asset="${asset_name}" '
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
  ' <<< "${release_json}"
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
  # Expand path in the trap string now — local tmp_dir is gone when EXIT runs.
  trap 'rm -rf "'"${tmp_dir}"'"' EXIT

  log "downloading ${ASSET_NAME} from ${RELEASE_TAG}"
  curl -fsSL "${ASSET_URL}" -o "${tmp_dir}/${ASSET_NAME}"
  chmod +x "${tmp_dir}/${ASSET_NAME}"

  DOWNLOADED_BINARY="${tmp_dir}/${ASSET_NAME}"

  if [[ -n "${CHECKSUMS_URL:-}" ]]; then
    verify_checksum
  else
    log "checksums.txt not found in release; skipping verification"
  fi
}

verify_checksum() {
  local checksum_file expected_hash actual_hash

  checksum_file="$(dirname "${DOWNLOADED_BINARY}")/checksums.txt"

  curl -fsSL "${CHECKSUMS_URL}" -o "${checksum_file}"

  # Strip \r in case checksums.txt was generated with CRLF line endings
  # (e.g. produced on Windows); otherwise $2 never matches asset.
  expected_hash="$(
    awk -v asset="${ASSET_NAME}" '{ sub(/\r$/, "") } $2 == asset { print $1; exit }' "${checksum_file}"
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

install_binary_to_path() {
  local target_path="$1"
  local target_dir old_path temp_target

  target_dir="$(dirname "${target_path}")"
  old_path="${target_path}.old"
  temp_target="${target_dir}/.$(basename "${target_path}").new"

  mkdir -p "${target_dir}"
  rm -f "${old_path}" "${temp_target}"

  install -m 0755 "${DOWNLOADED_BINARY}" "${temp_target}"

  if [[ -e "${target_path}" ]]; then
    mv -f "${target_path}" "${old_path}"
  fi

  if ! mv -f "${temp_target}" "${target_path}"; then
    if [[ -e "${old_path}" ]]; then
      mv -f "${old_path}" "${target_path}"
    fi
    rm -f "${temp_target}"
    fail "cannot install to ${target_path}"
  fi

  rm -f "${old_path}" "${temp_target}"
}

install_to_system() {
  if [[ -w "${SYSTEM_INSTALL_DIR}" ]]; then
    install_binary_to_path "${SYSTEM_INSTALL_PATH}"
    INSTALLED_PATH="${SYSTEM_INSTALL_PATH}"
    INSTALL_SCOPE="system"
    return 0
  fi

  if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    local tmp_dir old_path temp_target

    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "'"${tmp_dir}"'"' RETURN
    install -m 0755 "${DOWNLOADED_BINARY}" "${tmp_dir}/${ASSET_NAME}"

    old_path="${SYSTEM_INSTALL_PATH}.old"
    temp_target="${SYSTEM_INSTALL_DIR}/.pablo.new"
    sudo rm -f "${old_path}" "${temp_target}"
    sudo install -m 0755 "${tmp_dir}/${ASSET_NAME}" "${temp_target}"

    if [[ -e "${SYSTEM_INSTALL_PATH}" ]]; then
      sudo mv -f "${SYSTEM_INSTALL_PATH}" "${old_path}"
    fi

    if ! sudo mv -f "${temp_target}" "${SYSTEM_INSTALL_PATH}"; then
      if [[ -e "${old_path}" ]]; then
        sudo mv -f "${old_path}" "${SYSTEM_INSTALL_PATH}"
      fi
      sudo rm -f "${temp_target}"
      fail "cannot install to ${SYSTEM_INSTALL_PATH}"
    fi

    sudo rm -f "${old_path}" "${temp_target}"
    INSTALLED_PATH="${SYSTEM_INSTALL_PATH}"
    INSTALL_SCOPE="system"
    return 0
  fi

  return 1
}

install_to_user() {
  install_binary_to_path "${USER_INSTALL_PATH}"
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
