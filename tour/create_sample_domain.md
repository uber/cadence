---
codecontent: createsampledomain
weight: 15
---

# Create the Sample Domain

All workflows and activities in Cadence run within a domain. They need to be registered with the domain before 
they can be used. Let's create the domain that will be used for all the samples in this tour.

Note that you don't need to create the domain every time you start the Cadence server. If you used *Ctrl+C* to 
shut down the server, the contents of the database remain intact. This means that when you start the server 
again with `docker-compose up`, the domain will still be there. However, if you shut down the Cadence server 
with `docker-compose down`, that will wipe out the data and you'll need to create the domain again.

# Using the CLI

The Cadence CLI can be used directly from the Docker image *ubercadence/cli* or by 
[building the CLI tool](https://github.com/uber/cadence/tree/master/tools/cli#how) locally (which requires 
cloning the server code). The samples in this tour will use the Docker image. If using a locally built 
version, replace `docker run --rm ubercadence/cli:master` with `cadence`.

# Environment variables

To keep the CLI commands short, we'll use some environment variables. These include the following:

- **CADENCE_CLI_ADDRESS** - host:port for Cadence frontend service, the default is for the local server
- **CADENCE_CLI_DOMAIN** - default workflow domain, so you don't need to specify `--domain`
- **USER** - this is assigned by default in Linux/Mac but not on Windows
 