#!/bin/sh

set -ex

make go-generate && make fmt && make lint && make copyright

# buildkite-git complains about "dubious ownership",
# i.e. user-who-created is different than user-who-statused.
#
# that's fine, just override it if it looks like we're in buildkite.
# elsewhere it's probably best to not touch this, and the path is likely wrong.
if [[ -n $BUILDKITE_AGENT_ID ]]; then
  git config --global --add safe.directory /cadence
fi

# intentionally capture stderr, so status-errors are also PR-failing
if [ -n "$(git status --porcelain >&1)" ]; then
  echo "There are changes after make go-generate && make fmt && make lint && make copyright"
  echo "Please rerun the command and commit the changes"
  git status --porcelain
  exit 1
fi
