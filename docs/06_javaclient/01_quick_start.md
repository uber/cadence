# Quick Start

This topic helps you install the Cadence service and implement a workflow.

## Install Cadence Service Locally

### Install docker

Follow the Docker installation instructions found here: https://docs.docker.com/engine/installation/

### Run Cadence Server Using Docker Compose

Download the Cadence docker-compose file:
```bash
> curl -O https://raw.githubusercontent.com/uber/cadence/master/docker/docker-compose.yml
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   675  100   675    0     0    958      0 --:--:-- --:--:-- --:--:--   958
> ls
docker-compose.yml
```
Start Cadence Service:
```bash
> docker-compose up
Creating network "quick_start_default" with the default driver
Pulling cadence (ubercadence/server:0.5.8)...
0.5.8: Pulling from ubercadence/server
db0035920883: Pull complete
82eed7f2d38e: Pull complete
f81e11a89e41: Pull complete
ae3538b1ae1c: Pull complete
23ddfb58e314: Pull complete
52a6bbeb81b5: Pull complete
a72c7949d8ac: Pull complete
1c3b1d477195: Pull complete
3312d4123248: Pull complete
5bbc95a38c5f: Pull complete
29176d1ce1ca: Pull complete
27ec3755f89c: Pull complete
0a5d2a29a5e5: Pull complete
Creating quick_start_statsd_1    ... done
Creating quick_start_cassandra_1 ... done
Creating quick_start_cadence_1   ... done
Creating quick_start_cadence-web_1 ... done
Attaching to quick_start_cassandra_1, quick_start_statsd_1, quick_start_cadence_1, quick_start_cadence-web_1
statsd_1       | *** Running /etc/my_init.d/00_regen_ssh_host_keys.sh...
statsd_1       | *** Running /etc/my_init.d/01_conf_init.sh...
cadence_1      | + CADENCE_HOME=/cadence
cadence_1      | + DB=cassandra
...
...
...
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.199Z","msg":"Get dynamic config","name":"matching.longPollExpirationInterval","value":"1m0s","default-value":"1m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.199Z","msg":"Get dynamic config","name":"matching.updateAckInterval","value":"1m0s","default-value":"1m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.199Z","msg":"Get dynamic config","name":"matching.idleTasklistCheckInterval","value":"5m0s","default-value":"5m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.765Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Starting","wf-task-list-name":"cadence-archival-tl","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:185"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.775Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Started","wf-task-list-name":"cadence-archival-tl","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:199"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.891Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Starting","wf-task-list-name":"51f3b9fdfa7d:7feebe1f-95b2-44b8-8633-5ba7f4113508","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:185"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.900Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Started","wf-task-list-name":"51f3b9fdfa7d:7feebe1f-95b2-44b8-8633-5ba7f4113508","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:199"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:52.282Z","msg":"Get dynamic config","name":"history.shardUpdateMinInterval","value":"5m0s","default-value":"5m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:52.282Z","msg":"Get dynamic config","name":"history.emitShardDiffLog","value":"false","default-value":"false","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:27:24.903Z","msg":"Get dynamic config","name":"history.transferProcessorCompleteTransferFailureRetryCount","value":"10","default-value":"10","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:27:24.905Z","msg":"Get dynamic config","name":"history.timerProcessorCompleteTimerFailureRetryCount","value":"10","default-value":"10","logging-call-at":"config.go:57"}
```
### Register a Domain Using the CLI
From a different console window:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain domain register -rd 1
Unable to find image 'ubercadence/cli:master' locally
master: Pulling from ubercadence/cli
22dc81ace0ea: Pull complete
1a8b3c87dba3: Pull complete
91390a1c435a: Pull complete
07844b14977e: Pull complete
b78396653dae: Pull complete
5259e0c8568e: Pull complete
be8b5313e7cd: Pull complete
da2cfe74be81: Pull complete
5320bde81c0c: Pull complete
Digest: sha256:f5e5e708347909c8d3f74c47878b201d91606994394e94eaede9a80e3b9f077b
Status: Downloaded newer image for ubercadence/cli:master
Domain test-domain successfully registered.
>
```
Check that the domain is indeed registered:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain domain describe
Name: test-domain
Description:
OwnerEmail:
DomainData: map[]
Status: REGISTERED
RetentionInDays: 1
EmitMetrics: false
ActiveClusterName: active
Clusters: active
ArchivalStatus: DISABLED
Bad binaries to reset:
+-----------------+----------+------------+--------+
| BINARY CHECKSUM | OPERATOR | START TIME | REASON |
+-----------------+----------+------------+--------+
+-----------------+----------+------------+--------+
>
```
## Implement Hello World Java Workflow

### Include Cadence Java Client Dependency

Go to the [Maven Repository Uber Cadence Java Client Page](https://mvnrepository.com/artifact/com.uber.cadence/cadence-client)
and find the latest version of the library. Include it as a dependency into your Java project. For example if you
are using Gradle the dependency looks like:
```
compile group: 'com.uber.cadence', name: 'cadence-client', version: '<latest_version>'
```
Also add the following dependencies that cadence-client relies on:
```
    compile group: 'commons-configuration', name: 'commons-configuration', version: '1.9'
    compile group: 'ch.qos.logback', name: 'logback-classic', version: '1.2.3'
```
Make sure that the following code compiles:
```java
import com.uber.cadence.workflow.Workflow;
import com.uber.cadence.workflow.WorkflowMethod;
import org.slf4j.Logger;

public class GettingStarted {

    private static Logger logger = Workflow.getLogger(GettingStarted.class);

    interface HelloWorld {
        @WorkflowMethod
        void sayHello(String name);
    }

}
```
If you are having problems setting up the build files use the
[Cadence Java Samples](https://github.com/uber/cadence-java-samples) GitHub repository as a reference.

Also add the following logback config file somewhere in your classpath:
```xml
<configuration>
    <appender name="STDOUT" class="ch.qos.logback.core.ConsoleAppender">
        <!-- encoders are assigned the type
             ch.qos.logback.classic.encoder.PatternLayoutEncoder by default -->
        <encoder>
            <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>
    <logger name="io.netty" level="INFO"/>
    <root level="INFO">
        <appender-ref ref="STDOUT" />
    </root>
</configuration>
```

### Implement Hello World Workflow

Let's add `HelloWorldImpl` with the `sayHello` method that just logs the "Hello ..." and returns.
```java
import com.uber.cadence.worker.Worker;
import com.uber.cadence.workflow.Workflow;
import com.uber.cadence.workflow.WorkflowMethod;
import org.slf4j.Logger;

public class GettingStarted {

    private static Logger logger = Workflow.getLogger(GettingStarted.class);

    public interface HelloWorld {
        @WorkflowMethod
        void sayHello(String name);
    }

    public static class HelloWorldImpl implements HelloWorld {

        @Override
        public void sayHello(String name) {
            logger.info("Hello " + name + "!");
        }
    }
}
```
To link the workflow implementation to the Cadence framework, it should be registered with a worker that connects to
a Cadence Service. By default the worker connects to the locally running Cadence service.
```java
    public static void main(String[] args) {
        Worker.Factory factory = new Worker.Factory("test-domain");
        Worker worker = factory.newWorker("HelloWorldTaskList");
        worker.registerWorkflowImplementationTypes(HelloWorldImpl.class);
        factory.start();
    }
```
### Execute Hello World Workflow using the CLI

Now run the worker program. Following is an example log:
```text
13:35:02.575 [main] INFO  c.u.c.s.WorkflowServiceTChannel - Initialized TChannel for service cadence-frontend, LibraryVersion: 2.2.0, FeatureVersion: 1.0.0
13:35:02.671 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='Workflow Poller taskList="HelloWorldTaskList", domain="test-domain", type="workflow"'}, identity=45937@maxim-C02XD0AAJGH6}
13:35:02.673 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='null'}, identity=81b8d0ac-ff89-47e8-b842-3dd26337feea}
```
No Hello printed. This is expected because a worker is just a workflow code host. The workflow has to be started to execute. Let's use Cadence CLI to start the workflow:
```bash
> docker docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"World\"
Started Workflow Id: bcacfabd-9f9a-46ac-9b25-83bcea5d7fd7, run Id: e7c40431-8e23-485b-9649-e8f161219efe
```
The output of the program should change to:
```text
13:35:02.575 [main] INFO  c.u.c.s.WorkflowServiceTChannel - Initialized TChannel for service cadence-frontend, LibraryVersion: 2.2.0, FeatureVersion: 1.0.0
13:35:02.671 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='Workflow Poller taskList="HelloWorldTaskList", domain="test-domain", type="workflow"'}, identity=45937@maxim-C02XD0AAJGH6}
13:35:02.673 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='null'}, identity=81b8d0ac-ff89-47e8-b842-3dd26337feea}
13:40:28.308 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - Hello World!
```
Let's start another workflow execution:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"Cadence\"
Started Workflow Id: d2083532-9c68-49ab-90e1-d960175377a7, run Id: 331bfa04-834b-45a7-861e-bcb9f6ddae3e
```
And the output changed to:
```text
13:35:02.575 [main] INFO  c.u.c.s.WorkflowServiceTChannel - Initialized TChannel for service cadence-frontend, LibraryVersion: 2.2.0, FeatureVersion: 1.0.0
13:35:02.671 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='Workflow Poller taskList="HelloWorldTaskList", domain="test-domain", type="workflow"'}, identity=45937@maxim-C02XD0AAJGH6}
13:35:02.673 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='null'}, identity=81b8d0ac-ff89-47e8-b842-3dd26337feea}
13:40:28.308 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - Hello World!
13:42:34.994 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - Hello Cadence!
```
### List Workflows and Workflow History

Let's list our workflows in the CLI:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow list
             WORKFLOW TYPE            |             WORKFLOW ID              |                RUN ID                | START TIME | EXECUTION TIME | END TIME
  HelloWorld::sayHello                | d2083532-9c68-49ab-90e1-d960175377a7 | 331bfa04-834b-45a7-861e-bcb9f6ddae3e | 20:42:34   | 20:42:34       | 20:42:35
  HelloWorld::sayHello                | bcacfabd-9f9a-46ac-9b25-83bcea5d7fd7 | e7c40431-8e23-485b-9649-e8f161219efe | 20:40:28   | 20:40:28       | 20:40:29
```
Now let's look at the workflow execution history:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow showid 1965109f-607f-4b14-a5f2-24399a7b8fa7
  1  WorkflowExecutionStarted    {WorkflowType:{Name:HelloWorld::sayHello},
                                  TaskList:{Name:HelloWorldTaskList},
                                  Input:["World"],
                                  ExecutionStartToCloseTimeoutSeconds:3600,
                                  TaskStartToCloseTimeoutSeconds:10,
                                  ContinuedFailureDetails:[],
                                  LastCompletionResult:[],
                                  Identity:cadence-cli@linuxkit-025000000001,
                                  Attempt:0,
                                  FirstDecisionTaskBackoffSeconds:0}
  2  DecisionTaskScheduled       {TaskList:{Name:HelloWorldTaskList},
                                  StartToCloseTimeoutSeconds:10,
                                  Attempt:0}
  3  DecisionTaskStarted         {ScheduledEventId:2,
                                  Identity:45937@maxim-C02XD0AAJGH6,
                                  RequestId:481a14e5-67a4-436e-9a23-7f7fb7f87ef3}
  4  DecisionTaskCompleted       {ExecutionContext:[],
                                  ScheduledEventId:2,
                                  StartedEventId:3,
                                  Identity:45937@maxim-C02XD0AAJGH6}
  5  WorkflowExecutionCompleted  {Result:[],
                                  DecisionTaskCompletedEventId:4}
```
Even for such a trivial workflow, the history gives a lot of useful information. For complex workflows this is a really useful tool for production and development troubleshooting. History can be automatically archived to a long-term blob store (for example Amazon S3) upon workflow completion for compliance, analytical, and troubleshooting purposes.

### Workflow ID Uniqueness

Before proceeding to a more complex workflow implementation, let's take a look at the workflow ID semantic.
When starting a workflow without providing an ID, the client generates one in the form of a UUID. In most real-life scenarios this is not a desired behavior. The business ID should be used instead. Here, we'll specify the ID when starting a workflow:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start  --workflow_id "HelloCadence1" --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"Cadence\"
Started Workflow Id: HelloCadence1, run Id: 75170c60-6d72-48c6-b509-7c9d9f25a8a8
```
Now the list operation is more meaningful as the WORKFLOW ID is our business ID:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow list
             WORKFLOW TYPE            |             WORKFLOW ID              |                RUN ID                | START TIME | EXECUTION TIME | END TIME
  HelloWorld::sayHello                | HelloCadence1                        | 75170c60-6d72-48c6-b509-7c9d9f25a8a8 | 21:04:46   | 21:04:46       | 21:04:46
```
Let's try to start workflow with the same ID:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start  --workflow_id "HelloCadence1" --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"Cadence\"
Error: Failed to create workflow.
Error Details: WorkflowExecutionAlreadyStartedError{Message: Workflow execution already finished successfully. WorkflowId: HelloCadence1, RunId: 75170c60-6d72-48c6-b509-7c9d9f25a8a8. Workflow ID reuse policy: allow duplicate workflow ID if last run failed., StartRequestId: 350a03ed-a11f-4959-a424-8ff7166ed457, RunId: 75170c60-6d72-48c6-b509-7c9d9f25a8a8}
('export CADENCE_CLI_SHOW_STACKS=1' to see stack traces)
```

Oops, Cadence doesn't let us create a workflow with the same ID. But there are use cases when it is desired. For example if there is a need to re-execute the workflow for a particular reason. This is achieved by specifying a special flag _Workflow ID Reuse Policy_. The value of 1 means `AllowDuplicate`:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start  --workflowidreusepolicy 1 --workflow_id "HelloCadence1" --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"Cadence\"
Started Workflow Id: HelloCadence1, run Id: 37a740e5-838c-4020-aed6-1111b0689c38
```
After the second start the workflow list is:
```bash
     WORKFLOW TYPE     |             WORKFLOW ID              |                RUN ID                | START TIME | EXECUTION TIME | END TIME
  HelloWorld::sayHello | HelloCadence1                        | 37a740e5-838c-4020-aed6-1111b0689c38 | 21:11:47   | 21:11:47       | 21:11:47
  HelloWorld::sayHello | HelloCadence1                        | 75170c60-6d72-48c6-b509-7c9d9f25a8a8 | 21:04:46   | 21:04:46       | 21:04:46
```
It might be clear why every workflow has two IDs: Workflow ID and Run ID. Because the Workflow ID can be reused, the Run ID uniquely identifies a particular run of a workflow. Run ID is system generated and cannot be controlled by client code.

Note that ID Reuse Policy applies only when previous the run of a workflow is completed.
Under no circumstances does Cadence allow more than one instance of an open workflow with the same ID.

### CLI Help

You might be asking how to discover that 1 means `AllowDuplicate`. It came from the help command:
```bash
> docker run --network=host --rm ubercadence/cli:master workflow help start
NAME:
   cadence workflow start - start a new workflow execution

USAGE:
   cadence workflow start [command options] [arguments...]

OPTIONS:
   --tasklist value, --tl value                TaskList
   --workflow_id value, --wid value, -w value  WorkflowID
   --workflow_type value, --wt value           WorkflowTypeName
   --execution_timeout value, --et value       Execution start to close timeout in seconds (default: 0)
   --decision_timeout value, --dt value        Decision task start to close timeout in seconds (default: 10)
   --cron value                                Optional cron schedule for the workflow. Cron spec is as following:
                                               ┌───────────── minute (0 - 59)
                                               │ ┌───────────── hour (0 - 23)
                                               │ │ ┌───────────── day of the month (1 - 31)
                                               │ │ │ ┌───────────── month (1 - 12)
                                               │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
                                               │ │ │ │ │
                                               * * * * *
   --workflowidreusepolicy value, --wrp value  Optional input to configure if the same workflow ID is allowed to be used for a new workflow execution. Available options: 0: AllowDuplicateFailedOnly, 1: AllowDuplicate, 2: RejectDuplicate (default: 0)
   --input value, -i value                     Optional input for the workflow, in JSON format. If there are multiple parameters, concatenate them and separate by a space.
   --input_file value, --if value              Optional input for the workflow from a JSON file. If there are multiple JSON, concatenate them and separate by a space or newline. Input from the file will be overwritten by input from the command line.
   --memo_key value                            Optional key of memo. If there are multiple keys, concatenate them and separate by space.
   --memo value                                Optional info that can be shown in list workflow, in JSON format. If there are multiple JSON, concatenate them and separate by a space. The order must be the same as memo_key.
   --memo_file value                           Optional info that can be listed in list workflow, from JSON format file. If there are multiple JSON, concatenate them and separate by a space or newline. The order must be same as memo_key.
```
## Signals

So far our workflow is not very interesting. Let's change it to listen on an external event and update state accordingly.
```java
  public interface HelloWorld {
    @WorkflowMethod
    void sayHello(String name);

    @SignalMethod
    void updateGreeting(String greeting);
  }

  public static class HelloWorldImpl implements HelloWorld {

    private String greeting = "Hello";

    @Override
    public void sayHello(String name) {
      int count = 0;
      while (!"Bye".equals(greeting)) {
        logger.info(++count + ": " + greeting + " " + name + "!");
        String oldGreeting = greeting;
        Workflow.await(() -> !Objects.equals(greeting, oldGreeting));
      }
      logger.info(++count + ": " + greeting + " " + name + "!");
    }
    
    @Override
    public void updateGreeting(String greeting) {
      this.greeting = greeting;
    }
  }
```
The workflow interface now has a new method annotated with @SignalMethod. It is a callback method that is invoked
every time a new signal of "HelloWorld::updateGreeting" is delivered to a workflow. The workflow interface can have only
one @WorkflowMethod which is a _main_ function of the workflow and as many signal methods as needed.

The updated workflow implementation demonstrates a few important Cadence concepts. The first is that workflow is stateful and can
have fields of any complex type. Another is that the `Workflow.await` function that blocks until the function it receives as a parameter evaluates to true. The condition is going to be evaluated only on workflow state changes, so it is not a busy wait in traditional sense.   
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start  --workflow_id "HelloSignal" --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"World\"
Started Workflow Id: HelloSignal, run Id: 6fa204cb-f478-469a-9432-78060b83b6cd
```
Program output:
```text
16:53:56.120 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 1: Hello World!
```
Let's send a signal using CLI:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloSignal" --name "HelloWorld::updateGreeting" --input \"Hi\"
Signal workflow succeeded.
```
Program output:
```text
16:53:56.120 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 1: Hello World!
16:54:57.901 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 2: Hi World!
```
Try sending the same signal with the same input again. Note that the output doesn't change. This happens because the await condition
doesn't unblock when it sees the same value. But a new greeting unblocks it:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloSignal" --name "HelloWorld::updateGreeting" --input \"Welcome\"
Signal workflow succeeded.
```
Program output:
```text
16:53:56.120 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 1: Hello World!
16:54:57.901 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 2: Hi World!
16:56:24.400 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 3: Welcome World!
```
Now shut down the worker and send the same signal again:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloSignal" --name "HelloWorld::updateGreeting" --input \"Welcome\"
Signal workflow succeeded.
```
Note that sending signals as well as starting workflows does not need a worker running. The requests are queued inside the Cadence service.

Now bring the worker back. Note that it doesn't log anything besides the standard startup messages.
This occurs because it ignores the queued signal that contains the same input as the current value of greeting.
Note that the restart of the worker didn't affect the workflow execution. It is still blocked on the same line of code as before the failure.
This is the most important feature of Cadence. The workflow code doesn't need to deal with worker failures at all. Its state is fully recovered to its current state that includes all the local variables and threads.

Let's look at the line where the workflow is blocked:
```bash
> docker run --network=host --rm ubercadence/cli:master --do test-domain workflow stack --workflow_id "Hello2"
Query result:
"workflow-root: (BLOCKED on await)
com.uber.cadence.internal.sync.SyncDecisionContext.await(SyncDecisionContext.java:546)
com.uber.cadence.internal.sync.WorkflowInternal.await(WorkflowInternal.java:243)
com.uber.cadence.workflow.Workflow.await(Workflow.java:611)
com.uber.cadence.samples.hello.GettingStarted$HelloWorldImpl.sayHello(GettingStarted.java:32)
sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)"
```
Yes, indeed the workflow is blocked on await. This feature works for any open workflow, greatly simplifying troubleshooting in production.
Let's complete the workflow by sending a signal with a "Bye" greeting:

```text
16:58:22.962 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 4: Bye World!
```
Note that the value of the count variable was not lost during the restart.

Also note that while a single worker instance is used for this
walkthrough, any real production deployment has multiple worker instances running. So any worker failure or restart does not delay any
workflow execution because it is just migrated to any other available worker.

## Query

So far we have learned that the workflow code is fault tolerant and can update its state in reaction to external events in the form of signals.
Cadence provides a query feature that supports synchronously returning any information from a workflow to an external caller.

Update the workflow code to:
```java
  public interface HelloWorld {
    @WorkflowMethod
    void sayHello(String name);

    @SignalMethod
    void updateGreeting(String greeting);

    @QueryMethod
    int getCount();
  }

  public static class HelloWorldImpl implements HelloWorld {

    private String greeting = "Hello";
    private int count = 0;

    @Override
    public void sayHello(String name) {
      while (!"Bye".equals(greeting)) {
        logger.info(++count + ": " + greeting + " " + name + "!");
        String oldGreeting = greeting;
        Workflow.await(() -> !Objects.equals(greeting, oldGreeting));
      }
      logger.info(++count + ": " + greeting + " " + name + "!");
    }

    @Override
    public void updateGreeting(String greeting) {
      this.greeting = greeting;
    }

    @Override
    public int getCount() {
      return count;
    }
  }
```
The new `getCount` method annotated with `@QueryMethod` was added to the workflow interface definition. It is allowed
to have multiple query methods per workflow interface.

The main restriction on the implementation of the query method is that it is not allowed to modify workflow state in any form.
It also is not allowed to block its thread in any way. It usually just returns a value derived from the fields of the workflow object.
Let's run the updated worker and send a couple signals to it:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start  --workflow_id "HelloQuery" --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"World\"
Started Workflow Id: HelloQuery, run Id: 1925f668-45b5-4405-8cba-74f7c68c3135
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloQuery" --name "HelloWorld::updateGreeting" --input \"Hi\"
Signal workflow succeeded.
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloQuery" --name "HelloWorld::updateGreeting" --input \"Welcome\"
Signal workflow succeeded.
```
The worker output:
```text
17:35:50.485 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 1: Hello World!
17:36:10.483 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 2: Hi World!
17:36:16.204 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 3: Welcome World!
```
Now let's query the workflow using the CLI:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow query --workflow_id "HelloQuery" --query_type "HelloWorld::getCount"
Query result as JSON:
3
```
One limitation of the query is that it requires a worker process running because it is executing callback code.
An interesting feature of the query is that it works for completed workflows as well. Let's complete the workflow by sending "Bye" and query it.
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloQuery" --name "HelloWorld::updateGreeting" --input \"Bye\"
Signal workflow succeeded.
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow query --workflow_id "HelloQuery" --query_type "HelloWorld::getCount"
Query result as JSON:
4
```
The Query method can accept parameters. This might be useful if only part of the workflow state should be returned.

## Activities

Having fault tolerant code that maintains state, updates it in reaction to external events, and supports querying is already very useful.
But in most practical applications, the workflow is expected to act upon the external world. Cadence supports such externally-facing code in the form of activities.

An activity is essentially a function that can execute any code like DB updates or service calls. The workflow is not allowed to
directly call any external APIs; it can do this only through activities. The workflow is essentially an orchestrator of activities.
Let's change our program to print the greeting from an activity on every change.

First let's define an activities interface and implement it:
```java
  public interface HelloWorldActivities {
    @ActivityMethod(scheduleToCloseTimeoutSeconds = 100)
    void say(String message);
  }
```
The `@ActivityMethod` annotation is not required, but `scheduleToCloseTimeoutSeconds` is required and annotation is a convenient way to specify it.
It is allowed to have multiple activities on a single interface.

Activity implementation is just a normal [POJO](https://en.wikipedia.org/wiki/Plain_old_Java_object).
The `out` stream is passed as a parameter to the constructor to demonstrate that the
activity object can have any dependencies. Examples of real application dependencies are database connections and service clients.
```java
  public class HelloWordActivitiesImpl implements HelloWorldActivities {
    private final PrintStream out;

    public HelloWordActivitiesImpl(PrintStream out) {
      this.out = out;
    }

    @Override
    public void say(String message) {
      out.println(message);
    }
  }
```
Let's create a separate main method for the activity worker. It is common to have a single worker that hosts both activities and workflows,
but here we keep them separate to demonstrate how Cadence deals with worker failures.
To make the activity implementation known to Cadence, register it with the worker:
```java
public class GettingStartedActivityWorker {

  public static void main(String[] args) {
    Worker.Factory factory = new Worker.Factory("test-domain");
    Worker worker = factory.newWorker("HelloWorldTaskList");
    worker.registerActivitiesImplementations(new HelloWordActivitiesImpl(System.out));
    factory.start();
  }
}
```
A single instance of an activity object is registered per activity interface type. This means that the activity implementation should be thread-safe since the activity method can be simultaneously called from multiple threads.

Let's modify the workflow code to invoke the activity instead of logging:
```java
  public static class HelloWorldImpl implements HelloWorld {

    private final HelloWorldActivities activities = Workflow.newActivityStub(HelloWorldActivities.class);
    private String greeting = "Hello";
    private int count = 0;

    @Override
    public void sayHello(String name) {
      while (!"Bye".equals(greeting)) {
        activities.say(++count + ": " + greeting + " " + name + "!");
        String oldGreeting = greeting;
        Workflow.await(() -> !Objects.equals(greeting, oldGreeting));
      }
      activities.say(++count + ": " + greeting + " " + name + "!");
    }

    @Override
    public void updateGreeting(String greeting) {
      this.greeting = greeting;
    }

    @Override
    public int getCount() {
      return count;
    }
  }
```
Activities are invoked through a stub that implements their interface. So an invocation is just a method call on an activity stub.

Now run the workflow worker. Do not run the activity worker yet. Then start a new workflow execution:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow start  --workflow_id "HelloActivityWorker" --tasklist HelloWorldTaskList --workflow_type HelloWorld::sayHello --execution_timeout 3600 --input \"World\"
Started Workflow Id: HelloActivityWorker, run Id: ff015637-b5af-43e8-b3f6-8b6c7b919b62
```
The workflow is started, but nothing visible happens. This is expected as the activity worker is not running. What are the options to understand the currently running workflow state?

The first option is look at the stack trace:
```text
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow stack  --workflow_id "HelloActivityWorker"
Query result as JSON:
"workflow-root: (BLOCKED on Feature.get)com.uber.cadence.internal.sync.CompletablePromiseImpl.get(CompletablePromiseImpl.java:71)
com.uber.cadence.internal.sync.ActivityStubImpl.execute(ActivityStubImpl.java:58)
com.uber.cadence.internal.sync.ActivityInvocationHandler.lambda$invoke$0(ActivityInvocationHandler.java:87)
com.uber.cadence.internal.sync.ActivityInvocationHandler$$Lambda$25/1816732716.apply(Unknown Source)
com.uber.cadence.internal.sync.ActivityInvocationHandler.invoke(ActivityInvocationHandler.java:94)
com.sun.proxy.$Proxy6.say(Unknown Source)
com.uber.cadence.samples.hello.GettingStarted$HelloWorldImpl.sayHello(GettingStarted.java:55)
sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
"
```
It shows that the workflow code is blocked on the "say" method of a Proxy object that implements the activity stub.
You can restart the workflow worker if you want to make sure that restarting it does not change that. It works for activities
of any duration. It is okay for the workflow code to block on an activity invocation for a month for example.

Another way to see what exactly happened in the workflow execution is to look at the workflow execution history:
```text
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow show  --workflow_id "HelloActivityWorker"
  1  WorkflowExecutionStarted  {WorkflowType:{Name:HelloWorld::sayHello},
                                TaskList:{Name:HelloWorldTaskList},
                                Input:["World"],
                                ExecutionStartToCloseTimeoutSeconds:3600,
                                TaskStartToCloseTimeoutSeconds:10,
                                ContinuedFailureDetails:[],
                                LastCompletionResult:[],
                                Identity:cadence-cli@linuxkit-025000000001,
                                Attempt:0,
                                FirstDecisionTaskBackoffSeconds:0}
  2  DecisionTaskScheduled     {TaskList:{Name:HelloWorldTaskList},
                                StartToCloseTimeoutSeconds:10,
                                Attempt:0}
  3  DecisionTaskStarted       {ScheduledEventId:2,
                                Identity:36234@maxim-C02XD0AAJGH6,
                                RequestId:ef645576-7cee-4d2e-9892-597a08b7b01f}
  4  DecisionTaskCompleted     {ExecutionContext:[],
                                ScheduledEventId:2,
                                StartedEventId:3,
                                Identity:36234@maxim-C02XD0AAJGH6}
  5  ActivityTaskScheduled     {ActivityId:0,
                                ActivityType:{Name:HelloWorldActivities::say},
                                TaskList:{Name:HelloWorldTaskList},
                                Input:["1: Hello World!"],
                                ScheduleToCloseTimeoutSeconds:100,
                                ScheduleToStartTimeoutSeconds:100,
                                StartToCloseTimeoutSeconds:100,
                                HeartbeatTimeoutSeconds:100,
                                DecisionTaskCompletedEventId:4}
```
The last event in the workflow history is `ActivityTaskScheduled`. It is recorded when workflow invoked the activity, but it wasn't picked up by an activity worker yet.

Another useful API is `DescribeWorkflowExecution` which, among other information, contains the list of outstanding activities:
```text
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow describe  --workflow_id "HelloActivityWorker"
{
  "ExecutionConfiguration": {
    "taskList": {
      "name": "HelloWorldTaskList"
    },
    "executionStartToCloseTimeoutSeconds": 3600,
    "taskStartToCloseTimeoutSeconds": 10,
    "childPolicy": "TERMINATE"
  },
  "WorkflowExecutionInfo": {
    "Execution": {
      "workflowId": "HelloActivityWorker",
      "runId": "ff015637-b5af-43e8-b3f6-8b6c7b919b62"
    },
    "Type": {
      "name": "HelloWorld::sayHello"
    },
    "StartTime": "2019-06-08T23:56:41Z",
    "CloseTime": "1970-01-01T00:00:00Z",
    "CloseStatus": null,
    "HistoryLength": 5,
    "ParentDomainID": null,
    "ParentExecution": null,
    "AutoResetPoints": {}
  },
  "PendingActivities": [
    {
      "ActivityID": "0",
      "ActivityType": {
        "name": "HelloWorldActivities::say"
      },
      "State": "SCHEDULED",
      "ScheduledTimestamp": "2019-06-08T23:57:00Z"
    }
  ]
}
```
Let's start the activity worker. It starts and immediately prints:
```text
1: Hello World!
```
Let's look at the workflow execution history:
```text
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow show  --workflow_id "HelloActivityWorker"
   1  WorkflowExecutionStarted  {WorkflowType:{Name:HelloWorld::sayHello},
                                TaskList:{Name:HelloWorldTaskList},
                                Input:["World"],
                                ExecutionStartToCloseTimeoutSeconds:3600,
                                TaskStartToCloseTimeoutSeconds:10,
                                ContinuedFailureDetails:[],
                                LastCompletionResult:[],
                                Identity:cadence-cli@linuxkit-025000000001,
                                Attempt:0,
                                FirstDecisionTaskBackoffSeconds:0}
   2  DecisionTaskScheduled     {TaskList:{Name:HelloWorldTaskList},
                                StartToCloseTimeoutSeconds:10,
                                Attempt:0}
   3  DecisionTaskStarted       {ScheduledEventId:2,
                                Identity:37694@maxim-C02XD0AAJGH6,
                                RequestId:1d7cba6d-98c8-41fd-91b1-c27dffb21c7f}
   4  DecisionTaskCompleted     {ExecutionContext:[],
                                ScheduledEventId:2,
                                StartedEventId:3,
                                Identity:37694@maxim-C02XD0AAJGH6}
   5  ActivityTaskScheduled     {ActivityId:0,
                                ActivityType:{Name:HelloWorldActivities::say},
                                TaskList:{Name:HelloWorldTaskList},
                                Input:["1: Hello World!"],
                                ScheduleToCloseTimeoutSeconds:300,
                                ScheduleToStartTimeoutSeconds:300,
                                StartToCloseTimeoutSeconds:300,
                                HeartbeatTimeoutSeconds:300,
                                DecisionTaskCompletedEventId:4}
   6  ActivityTaskStarted       {ScheduledEventId:5,
                                Identity:37784@maxim-C02XD0AAJGH6,
                                RequestId:a646d5d2-566f-4f43-92d7-6689139ce944,
                                Attempt:0}
   7  ActivityTaskCompleted     {Result:[], ScheduledEventId:5,
                                StartedEventId:6,
                                Identity:37784@maxim-C02XD0AAJGH6}
   8  DecisionTaskScheduled     {TaskList:{Name:maxim-C02XD0AAJGH6:fd3a85ed-752d-4662-a49d-2665b7667c8a},
                                StartToCloseTimeoutSeconds:10, Attempt:0}
   9  DecisionTaskStarted       {ScheduledEventId:8,
                                Identity:fd3a85ed-752d-4662-a49d-2665b7667c8a,
                                RequestId:601ef30a-0d1b-4400-b034-65b8328ad34c}
  10  DecisionTaskCompleted     {ExecutionContext:[],
                                ScheduledEventId:8,
                                StartedEventId:9,
                                Identity:37694@maxim-C02XD0AAJGH6}
```
_ActivityTaskStarted_ event is recorded when the activity task is picked up by an activity worker. The Identity field
contains the ID of the worker (you can set it to any value on worker startup).

_ActivityTaskCompleted_ event is recorded when activity completes. It contains the result of the activity execution.

Let's look at various failure scenarios. Modify activity task timeout:
```java
  public interface HelloWorldActivities {
    @ActivityMethod(scheduleToCloseTimeoutSeconds = 100)
    void say(String message);
  }

  public class HelloWordActivitiesImpl implements HelloWorldActivities {
    private final PrintStream out;

    public HelloWordActivitiesImpl(PrintStream out) {
      this.out = out;
    }

    @Override
    public void say(String message) {
      out.println(message);
    }
  }
```
(To be continued ...)