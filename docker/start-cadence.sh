#!/bin/bash

set -ex

CONFIG_TEMPLATE_PATH="${CONFIG_TEMPLATE_PATH:-/etc/cadence/config/config_template.yaml}"

dockerize -template $CONFIG_TEMPLATE_PATH:/etc/cadence/config/docker.yaml

exec cadence-server --root $CADENCE_HOME --env docker start --services=$SERVICES
