#!/usr/bin/env bash
set -euo pipefail

if [ ! -f "bin/niskava" ]; then
    echo "Niskava executable not found. Compiling bin/niskava..."
    mkdir -p bin
    go build -o bin/niskava ./cmd/niskava
fi

exec ./bin/niskava "$@"
