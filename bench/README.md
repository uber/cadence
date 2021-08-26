Cadence Bench Tests
===================

This README describes how to set up Cadence bench, different types of bench loads, and how to start the load.

Setup
-----------
### Cadence server

Bench suite is running against a Cadence server/cluster. 

Note that only the Basic bench test don't require Advanced Visibility. 
 
Other advanced bench tests requires Cadence server with Advanced Visibility. 

For local env you can run it through:
- Docker: Instructions for running Cadence server through docker can be found in `docker/README.md`. Either `docker-compose-es-v7.yml` or `docker-compose-es.yml` can be used to start the server.
- Build from source: Please check [CONTRIBUTING](/CONTRIBUTING.md) for how to build and run Cadence server from source. Please also make sure Kafka and ElasticSearch are running before starting the server with `./cadence-server --zone es start`. If ElasticSearch v7 is used, change the value for `--zone` flag to `es_v7`.

See more [documentation here](https://cadenceworkflow.io/docs/concepts/search-workflows/).

### Bench Workers
:warning: NOTE: unlike canary, starting bench worker will not automatically start a bench test. Next two sections will cover how to start and configure it.

Different ways of start the bench workers:

#### 1. Use docker image `ubercadence/cadence-bench:latest`

For now, this image has on release versions for simplified the release process. Always use `latest` tag for the image. 

You can [pre-built docker-compose file](../docker/docker-compose-bench.yml) to run against local server
In the `docker/` directory, run:
```
docker-compose -f docker-compose-bench.yml up
```
You can modify [the bench worker config](../docker/config/bench/development.yaml) to run against a prod server cluster. 

Or may run it with Kubernetes, for [example](https://github.com/longquanzheng/cadence-lab/blob/master/eks/bench-deployment.yaml). 

  

#### 2.  Build & Run the binary 

In the project root, build cadence bench binary:
   ```
   make cadence-bench
   ```

Then start bench worker:
   ```
   ./cadence-bench start
   ```
By default, it will load [the configuration in `config/bench/development.yaml`](../config/bench/development.yaml). 
Run `./cadence-bench -h` for details to understand the start options of how to change the loading directory if needed. 

Worker Configurations
----------------------
Bench workers configuration contains two parts:
- **Bench**: this part controls the client side, including the bench service name, which domains bench workers are responsible for and how many taskLists each domain should use.
- **Cadence**: this control how bench worker should talk to Cadence server, which includes the server's service name and address.

Note:
1.  When starting bench workers, it will try to register a **local domain with archival feature disabled** for each domain name listed in the configuration, if not already exists. If your want to test the performance of global domains and/or archival feature, please register the domains first before starting the worker.
2.  Bench workers will only poll from task lists whose name start with `cadence-bench-tl-`. If in the configuration, `numTaskLists` is specified to be 2, then workers will only listen to `cadence-bench-tl-0` and `cadence-bench-tl-1`. So make sure you use a valid task list name when starting the bench load.

Bench Load Types
-----------
This section briefly describes the purpose of each bench load and provides a sample command for running the load. Detailed descriptions for each test's configuration can be found in `bench/lib/config.go`

Please note that all load configurations in `config/bench` is for only local development and illustration purpose, it does not reflect the actual capability of Cadence server.

### Basic
:warning: NOTE: This is the only bench test which doesn't require advanced visibility feature on the server. Make sure you set `useBasicVisibilityValidation` to true if run with basic(db) visibility.  
Also basicVisibilityValidation requires only one test load run in the same domain. This is because of the limitation of basic visibility now allow using workflowType and status filters at the same time.  

As the name suggests, this load tests the basic case of starting workflows and running activities in sequential/parallel. Once all test workflows are started, it will wait test workflow timeout + 5 mins before checking the status of all test workflows. If the failure rate is too high, or if there's any open workflows found, the test will fail.

The basic load can also be run in "panic" mode by setting `"panicStressWorkflow": true,` to test if server can handle large number of panic workflows (which can be caused by a bad worker deployment).

Sample configuration can be found in `config/bench/basic.json` and `config/ben/basic_panic.json`. To start the test, a sample command can be
```
cadence --do <domain> wf start --tl cadence-bench-tl-0 --wt basic-load-test-workflow --dt 30 --et 3600 --if config/bench/basic.json
```

`<domain>` needs to be one of the domains in bench config (by default  ./config/bench/development.yaml), e.g. `cadence-bench`. 

Then wait for the bench test result. 
```
$cadence --do cadence-bench wf ob -w a2813321-a1bd-40c6-934f-88ad0ded6037
Progress:
  1, 2021-08-20T11:49:14-07:00, WorkflowExecutionStarted
  2, 2021-08-20T11:49:14-07:00, DecisionTaskScheduled
...
...
  20, 2021-08-20T11:59:24-07:00, DecisionTaskStarted
  21, 2021-08-20T11:59:24-07:00, DecisionTaskCompleted
  22, 2021-08-20T11:59:24-07:00, WorkflowExecutionCompleted

Result:
  Run Time: 26 seconds
  Status: COMPLETED
  Output: "TEST PASSED: true; Details report: timeoutCount: 0, failedCount: 0, openCount:0, launchCount: 100, maxThreshold:1"

```
The test will return error if the test doesn't pass. There are two cases:
* The started stressWorkflow couldn't finish within the timeout
* There are more failed workflows than expected(`failureThreshold` * totalLaunchCount)

The output result is how many stressWorkflow were started successfully, and failed.

Configuration explnation
```
useBasicVisibilityValidation:   use basic(db based) visibility to verify the stress workflows, default false which requires advanced visibility on the server
totalLaunchCount	: total number of stressWorkflows that started by the launchWorkflow
routineCount	: number of in-parallel launch activities that started by launchWorkflow, to start the stressWorkflows
chainSequence	: number of steps in the stressWorkflow
concurrentCount	: number of in-parallel activity(dummy activity only echo data) in a step of the stressWorkflow
payloadSizeBytes	: payloadSize of echo data in the dummy activity
minCadenceSleepInSeconds	: control sleep time between two steps in the stressWorkflow, actual sleep time = random(min,max), default: 0
maxCadenceSleepInSeconds	: control sleep time between two steps in the stressWorkflow, actual sleep time = random(min,max), default: 0
executionStartToCloseTimeoutInSeconds	: StartToCloseTimeout of stressWorkflow, default 5m
contextTimeoutInSeconds	: RPC timeout for starting a stressWorkflow, default 3s
panicStressWorkflow	: if true, stressWorkflow will always panic, default false
failureThreshold	: the threshold of failed stressWorkflow for deciding whether or not the whole testSuite failed.
maxLauncherActivityRetryCount   : the max retry on launcher activity to start stress workflows, default: 5
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

### Cron: Run all the workloads as a TestSuite

:warning: NOTE: This requires a search attribute named `Passed` as boolean type. This search attribute should have been added to the [ES schema](/schema/elasticsearch). 
make sure the dynamic config also have [this search attribute (`frontend.validSearchAttributes`)](/config/dynamicconfig/development_es.yaml), so that Cadence server can recognize it.
* Validate `Passed` has been successfully added in the dynamic config:
   ```
   cadence cluster get-search-attr
   ```
   
`Cron` itself is not a test. It is responsible for running all other tests in parallel or sequential according a cron schedule. 

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