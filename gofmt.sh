#!/usr/bin/env bash

die () {
  >&2 echo 'could not build formatters, run `make fmt` for output'
  exit 1
}

# should match makefile's BIN var
BIN=.build/bin

# ensure tools exist
make -s "$BIN/gofmt" "$BIN/gofancyimports" || die

export PATH="$BIN:$PATH"
gofancyimports fix --local github.com/uber/cadence/ --write "$@"
gofmt -w "$@"
