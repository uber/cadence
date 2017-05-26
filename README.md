# Cadence 

Cadence is a distributed, scalable, durable, and highly available orchestration engine we developed at Uber Engineering to execute asynchronous long-running business logic in a scalable and resilient way.

This repo contains the source code of the Cadence server. Your application needs to use the client to interact with the server. The client can be found [here](https://github.com/uber-go/cadence-client).

## Running
### From source

* Build the binaries following the instructions [here](CONTRIBUTING.md).

* Install and run `cassandra` locally:
```bash
# for OS X
brew install cassandra

# start cassandra
/usr/local/bin/cassandra
```

* Setup the cassandra schema:
```bash
./cadence-cassandra-tool --ep 127.0.0.1 create -k "cadence" --rf 1
./cadence-cassandra-tool --ep 127.0.0.1 -k "cadence" setup-schema -d -f ./schema/cadence/schema.cql
```

* Start the service:
```bash
./cadence
```

### Docker

You can also [build and run](docker/README.md) the service using Docker. 

## Contributing
We'd love your help in making Cadence great. Please review our [instructions](CONTRIBUTING.md).

## License

MIT License, please see [LICENSE](https://github.com/uber/cadence/blob/master/LICENSE) for details.
