#!/bin/bash
set -e

cd "$(dirname "$0")/.."

go test -cover ./cmd/... ./pkg/...
go test -cover -coverpkg=./... ./tests/...