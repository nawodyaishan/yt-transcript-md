#!/usr/bin/env bash

source "$(dirname "$0")/lib/common.sh"

log_info "Building yt-transcript-md..."

VERSION=${VERSION:-dev}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}
DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}
GOVERSION=$(go version | awk '{print $3}')

go build -ldflags "-X github.com/nawodyaishan/yt-transcript-md/internal/version.Version=$VERSION \
                  -X github.com/nawodyaishan/yt-transcript-md/internal/version.Commit=$COMMIT \
                  -X github.com/nawodyaishan/yt-transcript-md/internal/version.Date=$DATE \
                  -X github.com/nawodyaishan/yt-transcript-md/internal/version.GoVersion=$GOVERSION" \
         -o bin/yt-transcript-md ./cmd/yt-transcript-md

log_info "Build complete: bin/yt-transcript-md"
