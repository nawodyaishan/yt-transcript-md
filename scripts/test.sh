#!/usr/bin/env bash

source "$(dirname "$0")/lib/common.sh"

log_info "Running tests..."

go test -v -coverprofile=coverage.out ./...

log_info "Tests complete."
