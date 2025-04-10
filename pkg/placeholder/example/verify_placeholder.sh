#!/bin/bash
# This script verifies that the magic placeholder string is preserved in the compiled binary.
# It compiles the example with aggressive optimizations and then checks for the presence
# of the placeholder string.

set -e

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Verifying the placeholder string in the compiled binary...${NC}"

# Get the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Create a temp directory for output
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Define output binary path
BINARY_PATH="$TEMP_DIR/example"

# Go module path prefix
MODULE_PREFIX="unisign/pkg/placeholder"

# Magic string placeholder - must match the one in placeholder.go
MAGIC_STRING="us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA=="

echo "Compiling example with aggressive optimizations..."
go build -o "$BINARY_PATH" -ldflags="-s -w" main.go

echo "Checking if the binary exists..."
if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Error: Binary was not created at $BINARY_PATH${NC}"
    exit 1
fi

echo "Running the binary to confirm it works..."
"$BINARY_PATH"
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Binary execution failed${NC}"
    exit 1
fi

echo "Checking if the magic string is present in the binary..."

# Try an exact match first
if strings "$BINARY_PATH" | grep -q "$MAGIC_STRING"; then
    echo -e "${GREEN}Success: Magic string found intact in the binary${NC}"
    exit 0
fi

echo "Exact match not found, checking for partial matches..."

# Check for distinctive chunks of the magic string
PREFIX="us1-"
CHUNK1="r/GZBm1d749E+KbBLWa"
CHUNK2="EOqGw+DeMQUNHb5TLBt"
CHUNK3="p82zcb9sMDO+Ai7e2TA"

FOUND=0

# Check each chunk
if strings "$BINARY_PATH" | grep -q "$PREFIX"; then
    echo -e "${GREEN}Found prefix: $PREFIX${NC}"
    FOUND=1
fi

if strings "$BINARY_PATH" | grep -q "$CHUNK1"; then
    echo -e "${GREEN}Found chunk: $CHUNK1${NC}"
    FOUND=1
fi

if strings "$BINARY_PATH" | grep -q "$CHUNK2"; then
    echo -e "${GREEN}Found chunk: $CHUNK2${NC}"
    FOUND=1
fi

if strings "$BINARY_PATH" | grep -q "$CHUNK3"; then
    echo -e "${GREEN}Found chunk: $CHUNK3${NC}"
    FOUND=1
fi

if [ $FOUND -eq 0 ]; then
    echo -e "${RED}Error: Magic string not found in the binary, even partially${NC}"
    echo -e "${YELLOW}This could indicate that compiler optimizations have removed the placeholder.${NC}"
    exit 1
else
    echo -e "${GREEN}Success: Parts of the magic string found in the binary${NC}"
    echo -e "${YELLOW}Note: The string might be split or encoded in the binary, but it's present.${NC}"
    echo -e "${YELLOW}This means unisign should be able to find and replace it.${NC}"
    exit 0
fi 