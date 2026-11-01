#!/bin/sh
# Run this after git tag to build the release for upload.

if ! git describe --exact-match HEAD; then
    echo HEAD is not a tag - create a tag first
    exit
fi

m() {
    local binary="zephyr.${GOOS}.${GOARCH}"
    [ "$GOOS" = "windows" ] && binary="${binary}.exe"
    go build -o "$binary"
    zstd "$binary"
}

GOOS=linux GOARCH=amd64 m

# Needs testing.
#GOOS=linux GOARCH=arm64 m
#GOOS=darwin GOARCH=arm64 m
#GOOS=windows GOARCH=amd64 m
#GOOS=windows GOARCH=arm64 m
