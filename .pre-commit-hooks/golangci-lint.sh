#!/usr/bin/env bash
set -e

# Try to find golangci-lint in common locations
if command -v golangci-lint > /dev/null 2>&1; then
    GOLANGCI_LINT="golangci-lint"
elif [ -f "/opt/homebrew/bin/golangci-lint" ]; then
    GOLANGCI_LINT="/opt/homebrew/bin/golangci-lint"
elif [ -f "/usr/local/bin/golangci-lint" ]; then
    GOLANGCI_LINT="/usr/local/bin/golangci-lint"
else
    echo "golangci-lint not found. Skipping lint check."
    exit 0
fi

cd code/cli
"$GOLANGCI_LINT" run ./...
