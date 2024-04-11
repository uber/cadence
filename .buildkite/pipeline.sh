#!/bin/bash

# exit immediately on failure, or if an undefined variable is used
set -eu

if [[ $BUILDKITE_BRANCH == 'master' || $BUILDKITE_BRANCH == 'tifflin/test-master-pipeline' ]]
then
    cat .buildkite/pipeline-master.yml
else
    cat .buildkite/pipeline-pull-request.yml
fi
