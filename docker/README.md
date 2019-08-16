# Quickstart for localhost development

Install docker: https://docs.docker.com/engine/installation/

Following steps will bring up the docker container running cadence server
along with all its dependencies (cassandra, statsd, graphite). Exposes cadence
frontend on port 7933 and grafana metrics frontend on port 8080.

```
$ cd $GOPATH/src/github.com/uber/cadence/docker
$ docker-compose up
```

View metrics at localhost:8080/dashboard    
View Cadence-Web at localhost:8088  
Use Cadence-CLI with `$ docker run --network=host --rm ubercadence/cli:master`

For example to register new domain 'test-domain' with 1 retention day
`$ docker run --network=host --rm ubercadence/cli:master --do test-domain domain register -rd 1`


## Using a pre-built image

With every tagged release of the cadence server, there is also a corresponding
docker image that's uploaded to docker hub. In addition, the release will also
contain a **docker.tar.gz** file (docker-compose startup scripts). 
Go [here](https://github.com/uber/cadence/releases/latest) to download a latest **docker.tar.gz** 

Execute the following
commands to start a pre-built image along with all dependencies (cassandra/statsd).

```
$ tar -xzvf docker.tar.gz
$ cd docker
$ docker-compose up
```

## Building an image for any branch and restarting

Replace **YOUR_TAG** and **YOUR_CHECKOUT_BRANCH** in the below command to build:
```
$ cd $GOPATH/src/github.com/uber/cadence
$ git checkout YOUR_CHECKOUT_BRANCH
$ docker build . -t ubercadence/server:YOUR_TAG --build-arg TARGET=auto-setup
```

Replace the tag of **image: ubercadence/server** to **YOUR_TAG** in docker-compose.yml .
Then stop service and remove all containers using the below commands.
```
$ docker-compose down
$ docker-compose up
```

## Running cadence service with MySQL

Run cadence with MySQL instead of Cassandra, use following commads:

```
$ docker-compose -f docker-compose-mysql.yml up
$ docker-compose -f docker-compose-mysql.yml down
```

Please note that SQL support is still in active developement and it is not production ready yet.

## Running cadence service with ElasticSearch

 Run cadence with ElasticSearch for visibility instead of Cassandra/MySQL

 ```
docker-compose -f docker-compose-es.yml up
``` 

# One host production setup with docker-compose

For a simple production setup, start with our **docker-compose-production-example.yml** for a single host Cadence cluster as described in the following steps.

## 1. Modify the database mounting path in the example YAML file
You may want to set the Cassandra data location because:
* a. You need a disk partition that has enough space
* b. Separate storage hardware for your database gives you better control over performance, availability, and backup strategy
* c. Setting a mounting path can avoid losing data when you run **docker-compose down**

Save the edited example YAML file as docker-compose.yaml.

## 2. Run docker-compose
```
$ docker-compose up -d
```
Try out our [samples](https://github.com/uber-common/cadence-samples) to make sure the server is running well.

See the [Docker Documentation](https://docs.docker.com/compose/production/) for more information about using docker-compose in production.

## 3. Upgrade Cadence server
* a. First, you need to upgrade the database schema. You can find instructions for using the cadence-cassandra-tool to upgrade schema [here](https://github.com/uber/cadence/tree/master/tools/cassandra#update-schema-as-part-of-a-release).
* b. Shut down docker-compose by using the following command:
```
$ docker-compose stop
```
* c. Modify the version tag in Cadence server to the new release in docker-compose.yaml. Then run it again:
```
$ docker-compose up -d
```

## 4. Backup
Use nodetool [snapshot](http://cassandra.apache.org/doc/latest/tools/nodetool/snapshot.html) command to create hot snapshots.
 To access nodetool, you can install nodetool, or log on to the Cassandra container. Get your Cassandra docker container ID by:
 ```
 $ docker ps
 692e6d194082        cassandra:3.11             "docker-entrypoint.s…"   14 minutes ago      Up 14 minutes       0.0.0.0:7199->7199/tcp, 7000-7001/tcp, 9160/tcp, 0.0.0.0:9042->9042/tcp                                             docker_cassandra_1
 ``` 
 Then 
 ```
 $ docker  exec -it 692e6d194082 /bin/bash
 ```
 Run the command to create snapshots (backups) for Cadence keyspaces: cadence and cadence-visibility.
 ```
 root@692e6d194082:/# nodetool  snapshot cadence cadence_visibility -t my-test-backup
 Requested creating snapshot(s) for [cadence, cadence_visibility] with snapshot name [my-test-backup] and options {skipFlush=false}
Snapshot directory: my-test-backup
 ```
 You will be able to see snapshots under each folder of Cadence's tables:
 For an example(commands are running under the cassandra data directory):
 ```
 $ ls data/cadence/domains-aec8fcd0a4ee11e98c689b89c16b65a4/snapshots/
my-test-backup
 ```
 You may need to copy the snapshots into other storage.
 To manage the snapshots use the listsnapshots and clearsnapshot commands. 
 Make sure to clean up the snapshots when you don't need them anymore, to save some space.
 
## 5. Restore
To restore from a backup (snapshot) for a table, using **domains** table as an example:
 
  First, make sure Cassandra is stopped.
  
  a. Clear the commitLogs:
   ```
   rm commitlog/*
   ```
  b. Go to the domains table directory and clear all data files (**NOT** folders):
  ```
  rm data/cadence/domains-aec8fcd0a4ee11e98c689b89c16b65a4/*.*
  ``` 
  c. Copy the snapshots to the table directory:
  ```
  cp data/cadence/domains-aec8fcd0a4ee11e98c689b89c16b65a4/snapshots/my-test-backup/* data/cadence/domains-aec8fcd0a4ee11e98c689b89c16b65a4/
  ```
  d. Finally, restart the Cassandra container and Cadence service.   

# Typical production setup

In a typical production setting, dependencies (Cassandra / MySQL and Statsd / Prometheus server) are
managed / started independently of the cadence-server. 
The recommended way is to use [Cadence Helm Chart](https://hub.helm.sh/charts/banzaicloud-stable/cadence), if you have Kubernetes.

 
To hook into other deployment platforms, you can use the following command:

```
$ docker run -e CASSANDRA_CONSISTENCY=Quorum \            -- Default cassandra consistency level
    -e CASSANDRA_SEEDS=10.x.x.x                         -- csv of cassandra server ipaddrs
    -e KEYSPACE=<keyspace>                              -- Cassandra keyspace
    -e VISIBILITY_KEYSPACE=<visibility_keyspace>        -- Cassandra visibility keyspace
    -e SKIP_SCHEMA_SETUP=true                           -- do not setup cassandra schema during startup
    -e RINGPOP_SEEDS=10.x.x.x,10.x.x.x  \               -- csv of ipaddrs for gossip bootstrap
    -e STATSD_ENDPOINT=10.x.x.x:8125                    -- statsd server endpoint
    -e NUM_HISTORY_SHARDS=1024  \                       -- Number of history shards
    -e SERVICES=history,matching \                      -- Spinup only the provided services
    -e LOG_LEVEL=debug,info \                           -- Logging level
    -e DYNAMIC_CONFIG_FILE_PATH=config/foo.yaml         -- Dynamic config file to be watched
    ubercadence/server:<tag>
```

