# Cadence
[![Build Status](https://badge.buildkite.com/159887afd42000f11126f85237317d4090de97b26c287ebc40.svg?theme=github&branch=master)](https://buildkite.com/uberopensource/cadence-server)
[![Coverage Status](https://coveralls.io/repos/github/uber/cadence/badge.svg)](https://coveralls.io/github/uber/cadence)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](http://t.uber.com/cadence-slack)

Visit [cadenceworkflow.io](https://cadenceworkflow.io) to learn about Cadence.

This repo contains the source code of the Cadence server. To implement workflows, activities and worker use [Go client](https://github.com/uber-go/cadence-client) or [Java client](https://github.com/uber-java/cadence-client).

See Maxim's talk at [Data@Scale Conference](https://atscaleconference.com/videos/cadence-microservice-architecture-beyond-requestreply) for an architectural overview of Cadence.

## Getting Started

### Start the cadence-server locally

We highly recommend that you use [Cadence service docker](docker/README.md) to run the service.

### Run the Samples

Try out the sample recipes for [Go](https://github.com/uber-common/cadence-samples) or [Java](https://github.com/uber/cadence-java-samples) to get started.

### Use CLI Tools
Use [Cadence command-line tool](tools/cli/README.md) to perform various tasks on Cadence server cluster

For [manual setup or upgrading](docs/persistence.md) server schema --

* If server runs with Cassandra, Use [Cadence Cassandra tool](tools/cassandra/README.md) to perform various tasks on database schema of Cassandra persistence
* If server runs with SQL database, Use [Cadence SQL tool](tools/sql/README.md) to perform various tasks on database schema of SQL based persistence

TIPS: Run `make tools` to build all tools mentioned above. 
> NOTE: See [CONTRIBUTING](docs/setup/CONTRIBUTING.md) for prerequisite of make command. 

### Use Cadence Web

Try out [Cadence Web UI](https://github.com/uber/cadence-web) to view your workflows on Cadence.
(This is already available at localhost:8088 if you run Cadence with docker compose)

### Cadence docs

If you have a suggestion or change you would like to make to [cadenceworkflow.io](https://cadenceworkflow.io) you can visit [Cadence Docs](https://github.com/uber/cadence-docs) and raise an Issue or Pull Request there.

## Contributing

We'd love your help in making Cadence great. Please review our [contribution guide](docs/setup/CONTRIBUTING.md).

If you'd like to propose a new feature, first join the Cadence [discussion group](https://groups.google.com/d/forum/cadence-discussion) and [Slack channel](https://join.slack.com/t/uber-cadence/shared_invite/enQtNDczNTgxMjYxNDEzLTQyYjcxZDM2YTIxMTZkMzQ0NjgxYmI3OWY5ODhiOTliM2I5MzA4NTM4MjU4YzgzZDkwNGEzOTUzNTBlNDk3Yjc) to start a discussion and check if there are existing design discussions. Also peruse our [design docs](docs/design/index.md) in case a feature has been designed but not yet implemented. Once you're sure the proposal is not covered elsewhere, please follow our [proposal instructions](PROPOSALS.md).

## License

MIT License, please see [LICENSE](https://github.com/uber/cadence/blob/master/LICENSE) for details.
