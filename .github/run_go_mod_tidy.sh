#!/bin/sh
# =============================================================================
#  This script updates Go modules to the latest version.
# =============================================================================
#  NOTE: This script is aimed to run in the container via docker-compose.
#    See "tidy" service: ./docker-compose.yml
# =============================================================================

set -eu

echo '* Backup modules ...'
cp go.mod go.mod.bak
cp go.sum go.sum.bak

echo '* Run go tidy ...'
go get -u ./...
go mod tidy -go=1.25.0

echo '* Run tests ...'
go test ./... && {
    echo '* Testing passed. Removing old go.mod file ...'
    rm -f go.mod.bak
    rm -f go.sum.bak
    echo 'Successfully updated modules!'
}
