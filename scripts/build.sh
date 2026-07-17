#!/usr/bin/env bash
set -euo pipefail

# Pablo Build Script
# This script handles multi-platform builds for the Pablo CLI.
# Usage: ./scripts/build.sh [all]
# Override output: BUILD_DIR=dist/releases ./scripts/build.sh all

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build directory is relative to the repo root unless absolute
BUILD_DIR="${BUILD_DIR:-build}"
case "${BUILD_DIR}" in
  /*) ;;
  *) BUILD_DIR="${ROOT}/${BUILD_DIR}" ;;
esac
mkdir -p "${BUILD_DIR}"

APP_NAME="pablo"

if [ "${1:-}" = "all" ]; then
    PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "windows/amd64" "windows/arm64")
    echo "Building for ALL platforms..."
else
    CURRENT_OS=$(go env GOOS)
    CURRENT_ARCH=$(go env GOARCH)
    PLATFORMS=("$CURRENT_OS/$CURRENT_ARCH")
    echo "Building for current platform ($CURRENT_OS/$CURRENT_ARCH)..."
fi

for PLATFORM in "${PLATFORMS[@]}"
do
    IFS="/" read -r OS ARCH <<< "$PLATFORM"

    if [ "${1:-}" = "all" ]; then
        OUTPUT_NAME="${APP_NAME}-${OS}-${ARCH}"
    else
        OUTPUT_NAME="${APP_NAME}"
    fi

    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo "--- Building for $OS/$ARCH..."

    # go.mod lives in src/; write artifacts under the resolved BUILD_DIR
    (cd "${ROOT}/src" && env GOOS=$OS GOARCH=$ARCH go build -o "${BUILD_DIR}/${OUTPUT_NAME}" .)
done

echo "----------------------------------------"
echo "Build complete! Check the '${BUILD_DIR}' directory."