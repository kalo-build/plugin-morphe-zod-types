@echo off
REM Build the Morphe Zod Types plugin as WASM

echo Building Morphe Zod Types plugin...

set GOOS=wasip1
set GOARCH=wasm

go build -o ./dist/morphe-zod-types-v1.0.0.wasm ./cmd/plugin/main.go

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Output: ./dist/morphe-zod-types-v1.0.0.wasm
) else (
    echo Build failed!
    exit /b 1
)
