# Using BuildKite

BuildKite simply runs Docker containers. So it is easy to perform the 
same build locally that BuildKite will do. To handle this, there are 
two different docker-compose files: one for BuildKite and one for local.
The Dockerfile is the same for both. 

## Testing the build locally
To try out the build locally, start from the root folder of this repo 
(cadence) and run the following commands.

Build the container for integration tests:
```bash
docker-compose -f docker/buildkite/docker-compose-local.yml build integrationtest
```

Run the integration tests:
```
docker-compose -f docker/buildkite/docker-compose-local.yml run integrationtest /bin/sh -e -c 'make cover_integration_ci'
```

Note that BuildKite will run basically the same commands.

## Testing the build in BuildKite
Creating a PR against the master branch will trigger the BuildKite
build. Members of the Cadence team can view the build pipeline here:
https://buildkite.com/uberopensource/cadence-server

Eventually this pipeline should be made public. It will need to ignore 
third party PRs for safety reasons.
