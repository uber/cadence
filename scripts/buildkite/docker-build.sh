#!/bin/bash

set -ex

echo "Building docker images for $BUILDKITE_MESSAGE"

docker build . -f Dockerfile -t ubercadence/server:master --build-arg TARGET=server
docker build . -f Dockerfile -t ubercadence/server:master-auto-setup --build-arg TARGET=auto-setup
docker build . -f Dockerfile -t ubercadence/server:master-auto-setup-with-kafka --build-arg TARGET=auto-setup-with-kafka
docker build . -f Dockerfile -t ubercadence/cli:master --build-arg TARGET=cli
docker build . -f Dockerfile -t ubercadence/cadence-bench:master --build-arg TARGET=bench
docker build . -f Dockerfile -t ubercadence/cadence-canary:master --build-arg TARGET=canary
