#!/bin/bash

# Target binary name
BINARY_NAME="pablo-lsp"

# Build platforms and architectures
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")

echo "Building Pablo LSP for multiple platforms..."

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS="/" read -r OS ARCH <<< "$PLATFORM"
    
    # Map OS to VS Code platform names
    VS_OS=$OS
    if [ "$OS" == "windows" ]; then VS_OS="win32"; fi

    EXT=""
    if [ "$OS" == "windows" ]; then EXT=".exe"; fi

    OUTPUT="../../extensions/pablo-lsp/build/${BINARY_NAME}-${VS_OS}-${ARCH}${EXT}"
    
    echo "  -> Building for ${OS}-${ARCH}..."
    GOOS=$OS GOARCH=$ARCH go build -o "$OUTPUT" main.go documents.go validator.go completion.go hover.go schema.go utils.go
done

echo "Build complete."
