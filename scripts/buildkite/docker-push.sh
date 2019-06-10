#!/bin/bash

set -ex

if [ "$BUILDKITE_BRANCH" != "master" ]; then
    echo "Skipping docker-push since this change is not in master yet"
    exit 0
fi

MASTER_SHA=$(git rev-parse origin/master)

if [ "$BUILDKITE_COMMIT" != "$MASTER_SHA" ]; then
    echo "Skipping docker-push for this commit since tip of master is already ahead"
    exit 0
fi

cd docker

echo "Building docker image for $BUILDKITE_MESSAGE"

docker build . -f Dockerfile -t ubercadence/server:master --build-arg git_branch=$BUILDKITE_COMMIT
docker push ubercadence/server:master

docker build . -f Dockerfile-cli -t ubercadence/cli:master --build-arg git_branch=$BUILDKITE_COMMIT 
docker push ubercadence/cli:master
