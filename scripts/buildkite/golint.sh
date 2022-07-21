#!/bin/sh

set -ex

make go-generate && make fmt && make lint && make copyright

if [ -n "$(git status --porcelain)" ]; then
  echo "There are changes after make go-generate && make fmt && make lint && make copyright"
  echo "Please rerun the command and commit the changes"
  git status --porcelain
  exit 1
fi
