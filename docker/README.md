Step #1: Build the image
========================
```
cd $GOPATH/src/github.com/uber/cadence
docker build -t 'uber/cadence:master' -f docker/Dockerfile docker/
```

Step #2: Run the image
======================

LocalHost Development
---------------------

```
docker run -p 7933:7933 -p 7934:7934 -p 7935:7935 uber/cadence:master
```

With all the config options
----------------------------
```
docker run -p 7933:7933 -p 7934:7934 -p 7935:7935
    -e NO_CASSANDRA=true \       -- Don't spin up cassandra as part of the docker
    -e CASSANDRA_SEEDS=192.168.0.1 \    -- talk to a remote cassandra server
    -e CASSANDRA_CONSISTENCY=Quorum \   -- Default cassandra consistency level
    -e BIND_ON_LOCALHOST=false \        -- Don't use localhost ip address for cadence services
    -e RINGPOP_SEEDS=10.0.0.1 \         -- Use this as the gossip bootstrap hosts for ringpop
    -e NUM_HISTORY_SHARDS=1024  \       -- Number of history shards
    -e SERVICES=history,matching \      -- Spinup only the provided services
    uber/cadence:master
```