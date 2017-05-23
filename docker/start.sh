#!/bin/bash -x

# Copyright (c) 2017 Uber Technologies, Inc.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

start_cassandra() {
    pushd /
    ./docker-entrypoint.sh cassandra
    popd

    until cqlsh "$HOST_IP" < /dev/null; do
        echo 'waiting for cassandra to start up'
        sleep 1
    done
    echo 'cassandra started'
}

setup_schema() {
    SCHEMA_FILE=$CADENCE_HOME/schema/cadence/schema.cql
    $CADENCE_HOME/cadence-cassandra-tool --ep $HOST_IP create -k $KEYSPACE --rf $RF
    $CADENCE_HOME/cadence-cassandra-tool --ep $HOST_IP -k $KEYSPACE setup-schema -d -f $SCHEMA_FILE
}

init_env() {

    export HOST_IP="127.0.0.1" # default to localhost binding
    export CASSANDRA_LISTEN_ADDRESS="127.0.0.1"

    if [ "$BIND_ON_LOCALHOST" == false ]; then
            export HOST_IP=`hostname --ip-address`
            export CASSANDRA_LISTEN_ADDRESS=$HOST_IP
    else
        export BIND_ON_LOCALHOST=true
    fi

    if [ -z "$KEYSPACE" ]; then
        export KEYSPACE="cadence"
    fi

    if [ -z "$CASSANDRA_SEEDS" ]; then
        export CASSANDRA_SEEDS=$HOST_IP
    fi

    if [ -z "$CASSANDRA_CONSISTENCY" ]; then
        export CASSANDRA_CONSISTENCY="One"
    fi

    if [ -z "$RINGPOP_SEEDS" ]; then
        export RINGPOP_SEEDS=$HOST_IP:7933,$HOST_IP:7934,$HOST_IP:7935
    fi

    if [ -z "$NUM_HISTORY_SHARDS" ]; then
        export NUM_HISTORY_SHARDS=4
    fi
}

CADENCE_HOME=$1

if [ -z "$RF" ]; then
    RF=1
fi

if [ -z "$SERVICES" ]; then
    SERVICES="history,matching,frontend"
fi

init_env

if [[ -z "$NO_CASSANDRA" || "$NO_CASSANDRA" == false ]]; then
    start_cassandra
    setup_schema
fi

# fix up config
envsubst < config/docker_template.yaml > config/docker.yaml
./cadence --root $CADENCE_HOME --env docker start --services=$SERVICES