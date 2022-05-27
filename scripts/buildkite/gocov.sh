#!/bin/sh

set -ex

# download cover files from all the tests
mkdir -p .build/coverage
buildkite-agent artifact download ".build/coverage/unit_cover.out" . --step ":golang: unit test" --build "$BUILDKITE_BUILD_ID"

echo "download complete"

# report coverage
make cover_ci

# cleanup
rm -rf .build
