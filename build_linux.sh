#!/bin/bash
set -e

echo "Building for Linux (amd64, static link)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o proxy-linux-amd64 cmd/proxy/main.go

echo "Done. Binary created at ./proxy-linux-amd64"
file proxy-linux-amd64
