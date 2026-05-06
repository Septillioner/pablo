#!/bin/bash

# Pablo Cross-Platform Publish Script (Unix)
# This script runs the Pablo deployment pipeline for the local system.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/src"

# Detect OS
OS_TYPE=$(uname | tr '[:upper:]' '[:lower:]')
ENV_NAME="${OS_TYPE}-local"

echo "========================================"
echo "  Pablo Self-Publishing ($OS_TYPE)"
echo "========================================"

sudo go run main.go -f ../pablo.yaml run -e "$ENV_NAME"
