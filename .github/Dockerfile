# =============================================================================
#  Unit Test Container for Vaious Go Versions (Last Updated: 2022-08-05)
# =============================================================================
# Default version: minimum supported Go version
ARG VARIANT="1.23-alpine"

# -----------------------------------------------------------------------------
#  Main Stage
# -----------------------------------------------------------------------------
FROM golang:${VARIANT}

RUN apk add --no-cache \
    git \
    alpine-sdk \
    build-base

WORKDIR /workspaces

ENTRYPOINT go mod download && go test -cover -race ./...
