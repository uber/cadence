## What
This package contains the tooling for cadence cassandra operations.

## How
- Run `make bins`
- You should see an executable `cadence-cassandra-tool`

## Setting up cassandra schema on a new cluster shortcut
``` 
make install-schema
```

## Setting up schema on a new cluster manually
```
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence setup-schema -v 0.0 -- this sets up just the schema version tables with initial version of 0.0
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence update-schema -d ./schema/cadence/versioned -- upgrades your schema to the latest version

./cadence-cassandra-tool -ep 127.0.0.1 -k cadence_visibility setup-schema -v 0.0 -- this sets up just the schema version tables with initial version of 0.0 for visibility
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence_visibility update-schema -d ./schema/visibility/versioned -- upgrades your schema to the latest version for visibility
```

## Updating schema on an existing cluster
You can only upgrade to a new version after the initial setup done above.

```
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence update-schema -d ./schema/cadence/versioned -v x.x -y -- executes a dryrun of upgrade to version x.x
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence update-schema -d ./schema/cadence/versioned -v x.x    -- actually executes the upgrade to version x.x

./cadence-cassandra-tool -ep 127.0.0.1 -k cadence_visibility update-schema -d ./schema/visibility/versioned -v x.x -y -- executes a dryrun of upgrade to version x.x
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence_visibility update-schema -d ./schema/visibility/versioned -v x.x    -- actually executes the upgrade to version x.x
```

## Using Docker 
There is a docker image for the `cadence-cassandra-tool` located at `ubercadence/cct`.  All the above commands could have the 
`./cadence-cassandra-tool` part replaced with `docker run --rm -it ubercadence/cct:<cadence version>`.

e.g. 
```
# old
./cadence-cassandra-tool -ep 127.0.0.1 -k cadence setup-schema -v 0.0

# new
docker run --rm -it ubercadence/cct:<cadence version> -ep 127.0.0.1 -k cadence setup-schema -v 0.0
```

The docker image also contains the schema migrations and the Makefile for the project.  So to setup the schema for a new cluster 
run:
```
docker run --rm -it --entrypoint /bin/bash ubercadence/cct:<cadence version> -c "make BUILD_BINS=false CASSANDRA_HOST=<cassandra host> CASSANDRA_REPLICATION_FACTOR=3 install-schema" 
```