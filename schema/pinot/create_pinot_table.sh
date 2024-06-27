#!/bin/bash

# Wait for Pinot components to start
sleep 30

# Add the table
/opt/pinot/bin/pinot-admin.sh AddTable \
  -schemaFile /schema/pinot/cadence-visibility-schema.json \
  -tableConfigFile /schema/pinot/cadence-visibility-config.json \
  -controllerProtocol http \
  -controllerHost pinot-controller \
  -controllerPort 9001 \
  -exec