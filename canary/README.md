# Periodical feature health check workflow tools(aka Canary)

This README describes how to set up Cadence bench, different types of bench loads, and how to start the load.

Setup
-----------
### Cadence server

Canary test suite is running against a Cadence server/cluster. See [documentation](https://cadenceworkflow.io/docs/operation-guide/setup/) for Cadence server cluster setup.

Note that some tests require features like [Advanced Visibility]((https://cadenceworkflow.io/docs/concepts/search-workflows/).) and [History Archival](https://cadenceworkflow.io/docs/concepts/archival/). 
 
For local server env you can run it through:
- Docker: Instructions for running Cadence server through docker can be found in `docker/README.md`. Either `docker-compose-es-v7.yml` or `docker-compose-es.yml` can be used to start the server.
- Build from source: Please check [CONTRIBUTING](/CONTRIBUTING.md) for how to build and run Cadence server from source. Please also make sure Kafka and ElasticSearch are running before starting the server with `./cadence-server --zone es start`. If ElasticSearch v7 is used, change the value for `--zone` flag to `es_v7`.

### Start canary worker 

:warning: NOTE: Starting this canary worker will not automatically start a canary test. Next two sections will cover how to start and configure it.

Different ways of start the bench workers:

#### 1. Use docker image `ubercadence/cadence-canary:master`

For now, this image has no release versions for simplified the release process. Always use `master` tag for the image. 

Similar to server/CLI images, the canary image will be built and published automatically by Github on every commit onto the `master` branch. 

You can [pre-built docker-compose file](../docker/docker-compose-canary.yml) to run against local server
In the `docker/` directory, run:
```
docker-compose -f docker-compose-canary.yml up
```
You can modify [the canary worker config](../docker/config/canary/development.yaml) to run against a prod server cluster. 


#### 2.  Build & Run the binary 

In the project root, build cadence canary binary:
 ```
 make cadence-canary
 ```

Then start canary worker & workflow:
 ```
 ./cadence-canary start
 ```
By default, it will load [the configuration in `config/canary/development.yaml`](../config/canary/development.yaml). 
Run `./cadence-canary -h` for details to understand the start options of how to change the loading directory if needed. 


Worker Configurations
----------------------
Canary workers configuration contains two parts:
- **Canary**: this part controls which domains canary workers are responsible for what tests the sanity workflow will exclude.
```yaml 
bench:
  domains: ["cadence-bench", "cadence-bench-sync", "cadence-bench-batch"] # it will start workers on all those domains(also try to register if not exists) 
  excludes: ["workflow.searchAttributes", "workflow.batch", "workflow.archival.visibility"] # it will exclude the three test cases
``` 
An exception here is `HistoryArchival` and `VisibilityArchival` test cases will always use `canary-archival-domain` domain. 

- **Cadence**: this control how bench worker should talk to Cadence server, which includes the server's service name and address.
```yaml
cadence:
  service: "cadence-frontend" # frontend service name
  host: "127.0.0.1:7933" # frontend address
```
- **Metrics**: metrics configuration. Similar to server metric emitter, only M3/Statsd/Prometheus is supported. 
- **Log**: logging configuration.  Similar to server logging configuration. 

Canary Test Cases
----------------------
#### Test case starter & Sanity suite 
The sanity workflow is test suite workflow. It will kick off a bunch of childWorkflows for all the test to verify that Cadence server is operating correctly. 

An error result of the sanity workflow indicates at least one of the test case fails.

You can start the sanity workflow as one-off run:
```
cadence --do <the domain you configured> workflow start --tl canary-task-queue --et 1200 --wt workflow.sanity -i 0
``` 
Note:
* tasklist(tl) is fixed to `canary-task-queue`
* execution timeout(et) is recommended to 20 minutes(`1200` seconds) but you can adjust it 
* the only required input is the scheduled unix timestamp, and `0` will uses the workflow starting time

Or using a cron job(e.g. every minute):
```
cadence --do <the domain you configured> workflow start --tl canary-task-queue --et 1200 --wt workflow.sanity -i 0 --cron "* * * * *"
```

This is [the list of the test cases](./sanity.go) that it will start all supported test cases by default if no excludes are configured. 
You can find [the workflow names of the tests cases in this file](./const.go) if you want to manually start certain test cases.  
For example, manually start an `Echo` test case:
```
cadence --do <> workflow start --tl canary-task-queue --et 10 --wt workflow.echo
```

#### Echo
Echo workflow tests the very basic workflow functionality. It executes an activity to return some output and verifies it as the workflow result. 

#### Signal
Signal workflow tests the signal feature. 

#### Visibility
Visibility workflow tests the basic visibility feature. No advanced visibility needed, but advanced visibility should also support it. 

#### SearchAttributes
SearchAttributes workflow tests the advanced visibility feature. Make sure advanced visibility feature is configured on the server. Otherwise, it should be excluded from the sanity test suite/case. 

#### ConcurrentExec
ConcurrentExec workflow tests executing activities concurrently. 

#### Query
Query workflow tests the Query feature. 

#### Timeout
Timeout workflow make sure the activity timeout is enforced. 

#### LocalActivity
LocalActivity workflow tests the local activity feature. 

#### Cancellation
Cancellation workflowt tests cancellation feature. 

#### Retry
Retry workflow tests activity retry policy. 

#### Reset
Reset workflow tests reset feature. 

#### HistoryArchival
HistoryArchival tests history archival feature. Make sure history archival feature is configured on the server. Otherwise, it should be excluded from the sanity test suite/case. 
This test case always uses `canary-archival-domain` domain.
 
#### VisibilityArchival
VisibilityArchival tests visibility archival feature. Make sure visibility feature is configured on the server. Otherwise, it should be excluded from the sanity test suite/case.

#### Batch  
Batch workflow tests the batch job feature. Make sure advanced visibility feature is configured on the server. Otherwise, it should be excluded from the sanity test suite/case.