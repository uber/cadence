There are 3 ways to create the tables in Pinot, need to make sure Pinot server is running. To start pinot server, use `docker compose -f ./docker/dev/cassandra-pinot-kafka.yml up`. 
Pinot has 4 components need to be started in order, so some components may fail to start for the first time. You can try again or restart the failed components in docker.

1. Manually add schema and realtime table in the UI, the configs can be found in the cadence-visibility-config file.

2. Use Pinot launcher script to add schema and table, this requires to download Apache Pinot.

`sh pinot-admin.sh AddTable \
-schemaFile [path_to_cadence]/cadence/Schema/Pinot/cadence-visibility-schema.json \
-tableConfigFile [path_to_cadence]/cadence/Schema/Pinot/cadence-visibility-config.json -exec`

3. Use docker to execute the launcher script.

`docker exec -it pinot-controller bin/pinot-admin.sh AddTable   \
-tableConfigFile /Schema/Pinot/cadence-visibility-schema.json   \
-schemaFile /Schema/Pinot/cadence-visibility-config.json -exec`

The result can be checked at http://localhost:9000/#/tables