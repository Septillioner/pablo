#!/bin/bash

# Pablo Build Script
# This script handles multi-platform builds for the Pablo CLI.

# Build directory
BUILD_DIR="build"
mkdir -p $BUILD_DIR

# App name
APP_NAME="pablo"

# Determine which platforms to build
if [ "$1" == "all" ]; then
    PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "windows/amd64" "windows/arm64")
    echo "Building for ALL platforms..."
else
    # Default to current platform
    CURRENT_OS=$(go env GOOS)
    CURRENT_ARCH=$(go env GOARCH)
    PLATFORMS=("$CURRENT_OS/$CURRENT_ARCH")
    echo "Building for current platform ($CURRENT_OS/$CURRENT_ARCH)..."
fi

for PLATFORM in "${PLATFORMS[@]}"
do
    # Split platform into OS and ARCH
    IFS="/" read -r OS ARCH <<< "$PLATFORM"
    
    # Set output filename
    if [ "$1" == "all" ]; then
        OUTPUT_NAME="${APP_NAME}-${OS}-${ARCH}"
    else
        # Default build name is just the app name
        OUTPUT_NAME="${APP_NAME}"
    fi

    # Append .exe for Windows
    if [ "$OS" == "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo "--- Building for $OS/$ARCH..."
    
    # Since go.mod is in src, we must run go build from the src directory
    (cd src && env GOOS=$OS GOARCH=$ARCH go build -o "../${BUILD_DIR}/${OUTPUT_NAME}" .)
done

echo "----------------------------------------"
echo "Build complete! Check the '${BUILD_DIR}' directory."