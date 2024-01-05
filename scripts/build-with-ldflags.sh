#!/bin/bash

set -eo pipefail

# add verbose env var for debugging
test -n "$V" && set -x

# must be on a separate line to trigger -e if it fails
FLAGS="$(./scripts/get-ldflags.sh LDFLAG)"
exec go build -ldflags "$FLAGS" "$@"
