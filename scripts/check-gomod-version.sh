#!/usr/bin/env bash
set -eo pipefail

if [[ $2 = "-v" ]]; then
    v () {
        >&2 echo "$@"
    }
else
    v () {
        true # noop
    }
fi

if ! gomod="$(go list -mod=readonly -f '{{ .Module }}' "$1")"; then
    >&2 echo 'Error checking main go.mod.'
    exit 1
fi

v "go.mod:                $gomod"

if ! toolmod="$(go list -mod=readonly -modfile=internal/tools/go.mod -f '{{ .Module }}' "$1")"; then
    >&2 echo 'Error checking tools go.mod, cd to internal/tools to modify it.'
    exit 1
fi

v "internal/tools/go.mod: $toolmod"

if [[ $gomod != $toolmod ]]; then
	>&2 echo "error: mismatched go.mod and tools go.mod"
    >&2 echo "ensure internal/tools/go.mod contains the same version as go.mod and try again:"
    >&2 echo -e "\t$gomod"
	exit 1
fi
