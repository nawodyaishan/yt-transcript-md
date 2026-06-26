#!/usr/bin/env bash

source "$(dirname "$0")/lib/common.sh"

if ! command -v docker &> /dev/null; then
    die "docker is not installed or not in PATH"
fi

log_info "Building Docker e2e test image..."
docker build -t yt-transcript-md-e2e tests/docker/

log_info "Running Docker e2e tests..."
go test -tags docker -v -timeout 180s ./tests/docker/...

log_info "Docker e2e tests complete."
