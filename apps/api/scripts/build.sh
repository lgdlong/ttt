#!/bin/sh
# Go build wrapper for Turborepo integration

# Set GOCACHE to a local directory if not set
if [ -z "$GOCACHE" ]; then
  export GOCACHE="$HOME/.cache/go-build"
fi

# Ensure the directory exists
mkdir -p "$GOCACHE"

# Run go build
exec go build -o tmp/main ./cmd/api/main.go
