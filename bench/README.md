Cadence Bench Tests
===================

This README describes how to set up Cadence bench, different types of bench loads, and how to start the load.

Setup
-----------
### Cadence server
Bench tests requires Cadence server with ElasticSearch. You can run it through:
- Docker: Instructions for running Cadence server through docker can be found in `docker/README.md`. Either `docker-compose-es-v7.yml` or `docker-compose-es.yml` can be used to start the server.
- Build from source: Please check [CONTRIBUTING](/CONTRIBUTING.md) for how to build and run Cadence server from source. Please also make sure Kafka and ElasticSearch are running before starting the server with `./cadence-server --zone es start`. If ElasticSearch v7 is used, change the value for `--zone` flag to `es_v7`.

### Search Attributes
One of the bench tests (called `Cron`), which is responsible for running other tests as a cron job and tracking the results, requires an search attribute named `Passed`. 

For local development environment, this search attribute has already been added to the ES index template and the list of valid search attributes.

However, if you already have a running ES cluster, you will need to add this search attribute to your ES cluster through the following steps:

1. Update ES cluster index template using the following Cadence CLI command
   ```
   cadence adm cluster asa --search_attr_key Passed --search_attr_type 4 
   ```
2. Add `Passed: 4` to the dynamic config value of valid search attributes (`frontend.validSearchAttributes`), so that Cadence server can recognize it.
3. Validate it has been successfully added with
   ```
   cadence cluster get-search-attr
   ```

### Bench Workers
For now there's no docker image for bench workers. The only way to run bench workers is:
1. Build cadence bench binary:
   ```
   make cadence-bench
   ```
2. Start bench workers:
   ```
   ./cadence-bench start
   ```
   By default, it will load the configuration in `config/bench/development.yaml`. Please run `./cadence-bench -h` for details on how to change the configuration directory and file used.
3. Note that, unlike canary, starting bench worker will not automatically start a bench test. Next two sections will cover how to start and configure it.

Worker Configurations
----------------------
Bench workers configuration contains two parts:
- **Bench**: this part controls the client side, including the bench service name, which domains bench workers are responsible for and how many taskLists each domain should use.
- **Cadence**: this control how bench worker should talk to Cadence server, which includes the server's service name and address.

Note:
1.  When starting bench workers, it will try to register a **local domain with archival feature disabled** for each domain name listed in the configuration, if not already exists. If your want to test the performance of global domains and/or archival feature, please register the domains first before starting the worker.
2.  Bench workers will only poll from task lists whose name start with `cadence-bench-tl-`. If in the configuration, `numTaskLists` is specified to be 2, then workers will only listen to `cadence-bench-tl-0` and `cadence-bench-tl-1`. So make sure you use a valid task list name when starting the bench load.

Bench Loads
-----------
This section briefly describes the purpose of each bench load and provides a sample command for running the load. Detailed descriptions for each test's configuration can be found in `bench/lib/config.go`

Please note that all load configurations in `config/bench` is for only local development and illustration purpose, it does not reflect the actual capability of Cadence server.

### Cron
`Cron` itself is not a test. It is responsible for running multiple other tests in parallel or sequential according a cron schedule. 

Tests in `Cron` are divided to into multiple test suites. Tests in different test suites will be run in parallel, while tests within a test suite will be run in a random sequential order. Different test suites can also be run in different domains, which provides a way for testing the multi-tenant performance of Cadence server. 

On the completion of each test, `Cron` will be signaled with the result of the test, which can be queried through:
```
cadence --do <domain> wf query --wid <workflowID of the Cron workflow> --qt test-results
```
This command will show the result of all completed tests.

When all tests complete, `Cron` will update the value of the `Passed` search attribute accordingly. `Passed` will be set to `true` only when all tests have passed, and `false` otherwise. Since the last event for cron workflow is always WorkflowContinuedAsNew, this search attribute can be used to tell whether one run of `Cron` is successful or not. You can see the search attribute value by adding `--psa` flag to workflow list commands when listing `Cron` runs.

A sample cron configuration is in `config/bench/cron.json`, and it can be started with
```
cadence --do <domain> wf start --tl cadence-bench-tl-0 --wt cron-test-workflow --dt 30 --et 7200 --if config/bench/cron.json
```

### Basic
As the name suggests, this load tests the basic case of starting workflows and running activities in sequential/parallel. Once all test workflows are started, it will wait test workflow timeout + 5 mins before checking the status of all test workflows. If the failure rate is too high, or if there's any open workflows found, the test will fail.

The basic load can also be run in "panic" mode by setting `"panicStressWorkflow": true,` to test if server can handle large number of panic workflows (which can be caused by a bad worker deployment).

Sample configuration can be found in `config/bench/basic.json` and `config/ben/basic_panic.json`. To start the test, a sample command can be
```
cadence --do <domain> wf start --tl cadence-bench-tl-0 --wt basic-load-test-workflow --dt 30 --et 3600 --if config/bench/basic.json
```

### Cancellation
The load tests the StartWorkflowExecution and CancelWorkflowExecution sync API, and validates the number of cancelled workflows and if there's any open workflow.

Sample configuration can be found in `config/bench/cancellation.json` and it can be started with
```
cadence --do <domain> wf start --tl cadence-bench-tl-0 --wt cancellation-load-test-workflow --dt 30 --et 3600 --if config/bench/cancellation.json 
```

### Signal
The load tests the SignalWorkflowExecution and SignalWithStartWorkflowExecution sync API, and validates the latency of signaling, the number of successfully completed workflows and if there's any open workflow.

Sample configuration can be found in `config/bench/signal.json` and it can be started with
```
cadence --do cadence-bench wf start --tl cadence-bench-tl-0 --wt signal-load-test-workflow --dt 30 --et 3600 --if config/bench/signal.json  
```


### Concurrent Execution
The purpose of this load is to test when a workflow schedules a large number of activities or child workflows in a single decision batch, whether server can properly throttle the processing of this workflow without affecting the execution of workflows in other domains. It will also check if the delayed period is within limit or not and fail the test if it takes too long.

A typical usage will be run this load and another load for testing sync APIs (for example, basic, cancellation or signal) in two different test suites/domains (so that they are run in parallel in two domains). Apply proper task processing throttling configuration to the domain that is running the concurrent execution test and see if tests in the other domain can still pass or not. 

Sample configuration can be found in `config/bench/concurrent_execution.json` and it can be started with
```
cadence --do <domain> wf start --tl cadence-bench-tl-0 --wt concurrent-execution-test-workflow --dt 30 --et 3600 --if config/bench/concurrent_execution.json
```

### Timer
This load tests if Cadence server can properly handle the case when one domain fires a large number of timers in a short period of time. Ideally timer from that domain should be throttled and delayed without affecting workflows in other domains. It will also check if the delayed period is within limit or not and fail the test if the timer latency is too high.

Typical usage is the same as the concurrent execution load above. Run it in parallel with another sync API test and see if the other test can pass or not.

Sample configuration can be found in `config/bench/timer.json` and it can be started with
```
cadence --do <domain> wf start --tl cadence-bench-tl-0 --wt timer-load-test-workflow --dt 30 --et 3600 --if config/bench/timer.json 
```