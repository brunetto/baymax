#!/usr/bin/env bash

version=`date -u +%Y%m%d-%H%M%S`

# Generate version reminder
rm last_version_number_*
touch "last_version_number_$version"

echo "Run go generate for rice box"
go generate

echo "Build package"
go build -ldflags "-X baymax.Version=$version"

echo "Cross-compile for linux (may take a while)"
GOOS=linux GOARCH=amd64 go build -ldflags "-X baymax.Version=$version"
