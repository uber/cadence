#!/bin/sh

set -ex

# fetch codecov reporting tool
go get github.com/dmetzgar/goveralls

# download cover files from all the tests
mkdir -p build/coverage
buildkite-agent artifact download "build/coverage/unit_cover.out" . --step ":golang: unit test" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download "build/coverage/integ_cassandra__cover.out" . --step ":golang: integration test with cassandra" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download "build/coverage/integ_ndc_cassandra__cover.out" . --step ":golang: integration ndc test with cassandra" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download "build/coverage/integ_sql_mysql_cover.out" . --step ":golang: integration test with mysql" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download "build/coverage/integ_ndc_sql_mysql_cover.out" . --step ":golang: integration ndc test with mysql" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download "build/coverage/integ_sql_postgres_cover.out" . --step ":golang: integration test with postgres" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download "build/coverage/integ_ndc_sql_postgres_cover.out" . --step ":golang: integration ndc test with postgres" --build "$BUILDKITE_BUILD_ID"


echo "download complete"

# report coverage
make cover_ci

# cleanup
rm -rf build
