# Releases upgrade instruction

## Upgrade to 0.16 and above

**TL;DR:** If you do not have workflows running for more than 6 months. It is safe to upgrade without any operations.

If your cluster has open workflows for more than 6 months, please download and run the following CLI command prior to the upgrades.

Cassandra:

`cadence admin db unsupported-workflow --db_type=cassandra --db_address <db address> --db_port <port number> --username=<username> --password=<password> --keyspace <keyspace> --number_of_shards=<your number of shards> --rps <scan rps> --output_filename ./cadence_scan`

MySQL/Postgres:

`cadence admin db unsupported-workflow --db_type=<mysql/postgres> --db_address <db address> --db_port <port number> --username=<username> --password=<password> --db_name <database name> --number_of_shards=<your number of shards> --rps <scan rps> --output_filename ./cadence_scan`

**Note:** This CLI is a long-running process to scan the database for unsupported workflows.
If you have TLS configurations or use customized encoding/decoding type. Please use

`cadence adm db unsupported-workflow --help`

to configurate the correct connection.


After the CLI completes, a list of unsupported workflows will be display in the file `cadence_scan`.

For example:

`cadence --address <host>:<port> --domain <ee7316ab-2e42-4f3c-86ee-344a4299791c> workflow reset --wid helloworld_7f2fa1fe-594e-44ad-a9e2-7ac7c5f97861 --rid e79ce972-b657-48a4-bba9-e2f2ac6424e2 --reset_type LastDecisionCompleted --reason 'release 0.16 upgrade'`

1. replace the host and port with your cadence host address and port number.
2. replace the domain uuid with the domain name.
To find the domain name, you can use `cadence --address <host>:<port> domain describe --domain_id=<domain uuid>`.

After replacing the fields with the correct values, you can copy/paste to run the CLI. This is going to reset those workflows to the last decision task completed event.

**Note:** If your workflows has outstanding child workflows, you have to wait for these child workflows completion or terminate the workflows.
