# cadence 

Cadence is a distributed, scalable, durable, and highly available orchestration engine we developed at Uber Engineering to execute asynchronous long-running business logic in a scalable and resilient way.

This repo contains the source code of the Cadence server. Your application needs to use the client to interact with the server. The client can be found [here](https://github.com/uber-go/cadence-client).

## Developing

### Build

Prerequisites:
* `thrift`: make certain that `thrift` is in your path. (OSX: `brew install thrift`) 
* `thrift-gen`: Go package must be present in your workspace (`go get github.com/uber/tchannel-go/thrift/thrift-gen`)
* `glide`: make sure `glide` is in your path. (https://glide.sh)

* Build the `cadence` service and other binaries (will not run test):
```
make bins
```

### Test

Prerequisites:
* all prereques listed for **build**
* `cassandra`: make sure you have `cassandra` 3.9 running locally and `cqlsh` is in your path (OSX: `brew install cassandra`)

```
make test
```

### Run the Cadence service

Prerequisites:
* you have already run `make bins`
* `cassandra`: make sure you have `cassandra` 3.9 running locally and `cqlsh` is in your path (OSX: `brew install cassandra`)

* Setup the cassandra schema:
```
./cadence-cassandra-tool --ep 127.0.0.1 create -k "cadence" --rf 1
./cadence-cassandra-tool --ep 127.0.0.1 -k "cadence" setup-schema -d -f ./schema/cadence/schema.cql
```

* Start the service:
```
./cadence
```

### Docker

You can also [build and run](docker/README.md) the service using Docker. 

## Contributing

We'd love your help in making Cadence great. If you need new API(s) to be added to our thrift files, open an issue and we will respond as fast as we can. If you want to propose new feature(s) along with the appropriate APIs yourself, open a pull request to start a discussion and we will merge it after review.

**Note:** All contributors also need to fill out the [Uber Contributor License Agreement](http://t.uber.com/cla) before we can merge in any of your changes

## Documentation

Coming soon ...

## License

MIT License, please see [LICENSE](https://github.com/uber/cadence/blob/master/LICENSE) for details.
