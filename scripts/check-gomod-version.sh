#!/usr/bin/env bash
set -eo pipefail

[[ $2 = "-v" ]] && set -x;

if ! gomod="$(go list -mod=readonly -f '{{ .Module }}' "$1")"; then
    >&2 echo 'Error checking main go.mod.'
    exit 1
fi

if ! toolmod="$(go list -mod=readonly -modfile=internal/tools/go.mod -f '{{ .Module }}' "$1")"; then
    >&2 echo 'Error checking tools go.mod, cd to internal/tools to modify it.'
    exit 1
fi

if [[ $gomod != $toolmod ]]; then
	>&2 echo "error: mismatched go.mod and tools go.mod"
    >&2 echo "ensure internal/tools/go.mod contains the same version as go.mod and try again:"
    >&2 echo -e "\t$gomod"
	exit 1
fi
