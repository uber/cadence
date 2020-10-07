#!/bin/bash

set -ex

dockerize -template /etc/cadence/config/config_template.yaml:/etc/cadence/config/docker.yaml

# Add read permission to non root users
chmod 644 /etc/cadence/config/* 
exec cadence-server --root $CADENCE_HOME --env docker start --services=$SERVICES
