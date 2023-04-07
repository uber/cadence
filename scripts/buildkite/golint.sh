#!/bin/sh

set -ex

if ! make go-generate; then
  >&2 echo 'failed to `make go-generate`, run it locally and commit any changes'
  exit 1
fi

if ! make lint; then
  >&2 echo 'failed to `make lint`, make sure lint passes'
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  >&2 echo "There are changes after make go-generate && make lint"
  >&2 echo "Please rerun the commands and fix and commit any changes"
  git status --porcelain
  exit 1
fi
