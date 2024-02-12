#!/bin/sh

set -ex

make tidy
make go-generate
make fmt
make lint
make copyright

# intentionally capture stderr, so status-errors are also PR-failing.
# in particular this catches "dubious ownership" failures, which otherwise
# do not fail this check and the $() hides the exit code.
if [ -n "$(git status --porcelain  2>&1)" ]; then
  echo "There are changes after make go-generate && make fmt && make lint && make copyright"
  echo "Please rerun the command and commit the changes"
  git status --porcelain
  git --no-pager diff
  exit 1
fi
