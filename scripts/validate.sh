#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo 'Running go fmt'
go fmt ./...

echo 'Running golangci-lint'
golangci-lint run

echo 'Tidying modules'
go mod tidy

echo 'Verifying modules'
go mod verify

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  echo "Encountered dirty repo!"
  exit 1
fi
