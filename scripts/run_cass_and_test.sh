#!/bin/bash

set -eo pipefail

function finish {
  echo "shut down containers"
  docker-compose -f docker/docker-compose.yml down;
}
trap finish EXIT

# shut down containers
docker-compose -f docker/docker-compose.yml down;

# run cassandra container
docker-compose -f docker/docker-compose.yml up -d cassandra;

# run the tests with cassandra
CASSANDRA=1 make test | tee test.log;
