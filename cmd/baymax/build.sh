#!/usr/bin/env bash

cd ../../
build.sh
cd cmd/baymax

version=`date -u +%Y%m%d-%H%M%S`
output="$(pwd)/${PWD##*/}"

# Generate version reminder
rm last_version_number_*
touch "last_version_number_$version"

echo "Build"
go build -o ${output} -ldflags "-X main.version=$version" main.go

echo "Cross-compile for linux (may take a while)"
GOOS=linux GOARCH=amd64 go build -o ${output}_linux -ldflags "-X main.version=$version" main.go
