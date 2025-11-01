#!/bin/bash

# Build the Morphe Zod Types plugin as WASM
echo "Building Morphe Zod Types plugin..."

GOOS=wasip1 GOARCH=wasm go build -o ./dist/morphe-zod-types-v1.0.0.wasm ./cmd/plugin/main.go

if [ $? -eq 0 ]; then
    echo "Build successful! Output: ./dist/morphe-zod-types-v1.0.0.wasm"
else
    echo "Build failed!"
    exit 1
fi
