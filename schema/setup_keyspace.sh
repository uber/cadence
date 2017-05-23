#!/bin/bash

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

set -xeo pipefail

# This script is used to setup the "cadence" keyspace
# and load all the tables in the .cql file within this keyspace,
# assuming cassandra is up and running.
# The default keyspace is "cadence", default replication factor is 3.

# To run against a specific Cassandra deployment, set CQLSH_HOST to the
# Cassandra IP address

if [ -z "$KEYSPACE" ]; then
    KEYSPACE="cadence"
fi

if [ -z "$RF" ]; then
    RF=3
fi

if [ -z "$CQLSH" ]; then
    CQLSH=cqlsh
fi

$CQLSH -e "CREATE KEYSPACE IF NOT EXISTS $KEYSPACE WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '$RF'};"
$CQLSH -k $KEYSPACE -f `dirname $0`/cadence/keyspace.cql