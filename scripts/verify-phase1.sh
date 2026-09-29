#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase1.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

go test ./...
go test -race ./...
go vet ./...
go run ./cmd/schway verify testdata/phase1
