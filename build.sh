#!/bin/bash

# CipherHub Build Script
# Usage: ./build.sh [target]
# Targets: windows, linux, darwin, all, clean

set -e

VERSION=${VERSION:-"1.0.0"}
BINARY_NAME="cipherhub"
MAIN_PATH="./cmd/cipherhub"
BUILD_DIR="bin"

mkdir -p $BUILD_DIR

build_windows() {
    echo "Building for Windows AMD64..."
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/$BINARY_NAME-windows-amd64.exe $MAIN_PATH
    echo "✓ Built: $BUILD_DIR/$BINARY_NAME-windows-amd64.exe"
}

build_windows_arm64() {
    echo "Building for Windows ARM64..."
    GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/$BINARY_NAME-windows-arm64.exe $MAIN_PATH
    echo "✓ Built: $BUILD_DIR/$BINARY_NAME-windows-arm64.exe"
}

build_linux() {
    echo "Building for Linux AMD64..."
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/$BINARY_NAME-linux-amd64 $MAIN_PATH
    echo "✓ Built: $BUILD_DIR/$BINARY_NAME-linux-amd64"
}

build_linux_arm64() {
    echo "Building for Linux ARM64..."
    GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/$BINARY_NAME-linux-arm64 $MAIN_PATH
    echo "✓ Built: $BUILD_DIR/$BINARY_NAME-linux-arm64"
}

build_darwin() {
    echo "Building for macOS AMD64..."
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/$BINARY_NAME-darwin-amd64 $MAIN_PATH
    echo "✓ Built: $BUILD_DIR/$BINARY_NAME-darwin-amd64"
}

build_darwin_arm64() {
    echo "Building for macOS ARM64 (M1/M2)..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/$BINARY_NAME-darwin-arm64 $MAIN_PATH
    echo "✓ Built: $BUILD_DIR/$BINARY_NAME-darwin-arm64"
}

clean() {
    echo "Cleaning build directory..."
    rm -rf $BUILD_DIR
    go clean
    echo "✓ Cleaned"
}

case "$1" in
    windows)
        build_windows
        ;;
    windows-arm64)
        build_windows_arm64
        ;;
    linux)
        build_linux
        ;;
    linux-arm64)
        build_linux_arm64
        ;;
    darwin)
        build_darwin
        ;;
    darwin-arm64)
        build_darwin_arm64
        ;;
    all)
        build_windows
        build_linux
        build_darwin
        build_darwin_arm64
        ;;
    clean)
        clean
        ;;
    *)
        echo "CipherHub Build Script"
        echo ""
        echo "Usage: $0 [target]"
        echo ""
        echo "Targets:"
        echo "  windows        Build for Windows AMD64"
        echo "  windows-arm64  Build for Windows ARM64"
        echo "  linux          Build for Linux AMD64"
        echo "  linux-arm64    Build for Linux ARM64"
        echo "  darwin         Build for macOS AMD64"
        echo "  darwin-arm64   Build for macOS ARM64 (M1/M2)"
        echo "  all            Build for all platforms"
        echo "  clean          Clean build directory"
        echo ""
        echo "Current default: Building Windows AMD64"
        build_windows
        ;;
esac
