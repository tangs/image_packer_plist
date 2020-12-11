#!/usr/bin/env sh

export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
go build src/png_packer.go
