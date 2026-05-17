#!/usr/bin/env bash

set -euo pipefail

# Color codes for output
export COLOR_RESET="\033[0m"
export COLOR_INFO="\033[32m"
export COLOR_ERROR="\033[31m"
export COLOR_WARN="\033[33m"

function log_info() {
    echo -e "${COLOR_INFO}[INFO]${COLOR_RESET} $1"
}

function log_warn() {
    echo -e "${COLOR_WARN}[WARN]${COLOR_RESET} $1"
}

function log_error() {
    echo -e "${COLOR_ERROR}[ERROR]${COLOR_RESET} $1"
}

function die() {
    log_error "$1"
    exit 1
}
