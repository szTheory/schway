#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/codename-lang-phase1.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

go test ./...
go test -race ./...
go vet ./...
go run ./cmd/lang verify testdata/phase1
