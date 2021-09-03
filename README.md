# Cadence
[![Build Status](https://badge.buildkite.com/159887afd42000f11126f85237317d4090de97b26c287ebc40.svg?theme=github&branch=master)](https://buildkite.com/uberopensource/cadence-server)
[![Coverage Status](https://coveralls.io/repos/github/uber/cadence/badge.svg)](https://coveralls.io/github/uber/cadence)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](http://t.uber.com/cadence-slack)

This repo contains the source code of the Cadence server and other tooling including CLI, schema tools, bench and canary. 

You can implement your workflows with one of our client libraries. 
The [Go](https://github.com/uber-go/cadence-client) and [Java](https://github.com/uber-java/cadence-client) libraries are officially maintained by the Cadence team, 
while the [Python](https://github.com/firdaus/cadence-python) and [Ruby](https://github.com/coinbase/cadence-ruby) client libraries are developed by the community.

See Maxim's talk at [Data@Scale Conference](https://atscaleconference.com/videos/cadence-microservice-architecture-beyond-requestreply) for an architectural overview of Cadence.

Visit [cadenceworkflow.io](https://cadenceworkflow.io) to learn more about Cadence. Join us in [Cadence Documentation](https://github.com/uber/cadence-docs) project. Feel free to raise an Issue or Pull Request there.

### Community 
* [Github Discussion](https://github.com/uber/cadence/discussions)
  * Best for Q&A, support/help, general discusion, and annoucement 
* [StackOverflow](https://stackoverflow.com/questions/tagged/cadence-workflow)
  * Best for Q&A and general discusion
* [Github Issues](https://github.com/uber/cadence/issues)
  * Best for reporting bugs and feature requests
* [Slack](http://t.uber.com/cadence-slack)
  * Best for contributing/development discussion 
  
## Getting Started

### Start the cadence-server

To run Cadence services locally, we highly recommend that you use [Cadence service docker](docker/README.md) to run the service.
You can also follow the [instructions](./CONTRIBUTING.md) to build and run it. 

Please visit our [documentation](https://cadenceworkflow.io/docs/operation-guide/) site for production/cluster setup.

### Run the Samples

Try out the sample recipes for [Go](https://github.com/uber-common/cadence-samples) or [Java](https://github.com/uber/cadence-java-samples) to get started.

### Use [Cadence CLI](https://cadenceworkflow.io/docs/cli/) to perform various tasks on Cadence server cluster

You can use the following ways to install Cadence CLI:
* Use brew to install CLI: `brew install cadence-workflow`
* Use docker image for CLI: `docker run --rm ubercadence/cli:<releaseVersion>`  or `docker run --rm ubercadence/cli:master ` . Be sure to update your image when you want to try new features: `docker pull ubercadence/cli:master `
* Build the CLI binary yourself, check out the repo and run `make cadence` to build all tools. See [CONTRIBUTING](CONTRIBUTING.md) for prerequisite of make command.
* Build the CLI image yourself, see [instructions](docker/README.md#diy-building-an-image-for-any-tag-or-branch)
  
Cadence CLI is a powerful tool. The commands are organized by **tabs**. E.g. `workflow`->`batch`->`start`, or `admin`->`workflow`->`describe`.

Please read the [documentation](https://cadenceworkflow.io/docs/cli/#documentation) to learn & explore.  
  
### Use Cadence Web

Try out [Cadence Web UI](https://github.com/uber/cadence-web) to view your workflows on Cadence.
(This is already available at localhost:8088 if you run Cadence with docker compose)


## Contributing

We'd love your help in making Cadence great. Please review our [contribution guide](CONTRIBUTING.md).

If you'd like to propose a new feature, first join the [Slack channel](http://t.uber.com/cadence-slack) to start a discussion and check if there are existing design discussions. Also peruse our [design docs](docs/design/index.md) in case a feature has been designed but not yet implemented. Once you're sure the proposal is not covered elsewhere, please follow our [proposal instructions](PROPOSALS.md).

## Other binaries in this repo

#### Bench/stress test workflow tools
See [bench documentation](./bench/README.md). 

#### Periodical feature health check workflow tools(aka Canary)
See [canary documentation](./canary/README.md).

#### Schema tools for SQL and Cassandra
The tools are for [manual setup or upgrading database schema](docs/persistence.md)  
 
  * If server runs with Cassandra, Use [Cadence Cassandra tool](tools/cassandra/README.md) 
  * If server runs with SQL database, Use [Cadence SQL tool](tools/sql/README.md) 

The easiest way to get the schema tool is via homebrew.

`brew install cadence-workflow` also includes `cadence-sql-tool` and `cadence-cassandra-tool`. 
 * The schema files are located at `/usr/local/etc/cadence/schema/`.
    
## License

MIT License, please see [LICENSE](https://github.com/uber/cadence/blob/master/LICENSE) for details.
