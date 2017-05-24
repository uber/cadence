Step #1: Build the image
========================
```
cd $GOPATH/src/github.com/uber/cadence/docker
docker-compose build
```

Step #2: Run the image
======================

LocalHost Development
---------------------

```
docker-compose run cadence
```

With all the config options
----------------------------
```
docker-compose run -e CASSANDRA_CONSISTENCY=Quorum \   -- Default cassandra consistency level
    -e BIND_ON_LOCALHOST=false \        -- Don't use localhost ip address for cadence services
    -e RINGPOP_SEEDS=10.0.0.1 \         -- Use this as the gossip bootstrap hosts for ringpop
    -e NUM_HISTORY_SHARDS=1024  \       -- Number of history shards
    -e SERVICES=history,matching \      -- Spinup only the provided services
    cadence
```