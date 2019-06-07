# Quick Start

## Install Cadence Service Locally

### Install docker

Docker installation instructions: https://docs.docker.com/engine/installation/

### Run Cadence Server Using Docker Compose

Download Cadence docker-compose file:
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
### Register Domain Using CLI
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

Go to [Maven Repository Uber Cadence Java Client Page](https://mvnrepository.com/artifact/com.uber.cadence/cadence-client)
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
If you are having problems setting up the build files use 
[Cadence Java Samples](https://github.com/uber/cadence-java-samples) Github repository as a reference.

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
Let's add HelloWorldImpl with sayHello method that just logs the "Hello ..." and returns.
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
To link the workflow implementation to Cadence framework it should be registered with a worker that connects to 
a Cadence Service. By default the worker connects to the locally running Cadence service.
```java
    public static void main(String[] args) {
        Worker.Factory factory = new Worker.Factory("test-domain");
        Worker worker = factory.newWorker("HelloWorldTaskList");
        worker.registerWorkflowImplementationTypes(HelloWorldImpl.class);
        factory.start();
    }
```
### Execute Hello World Workflow Through CLI
Now run the worker program. This is what it logs on my machine:
```text
13:35:02.575 [main] INFO  c.u.c.s.WorkflowServiceTChannel - Initialized TChannel for service cadence-frontend, LibraryVersion: 2.2.0, FeatureVersion: 1.0.0
13:35:02.671 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='Workflow Poller taskList="HelloWorldTaskList", domain="test-domain", type="workflow"'}, identity=45937@maxim-C02XD0AAJGH6}
13:35:02.673 [main] INFO  c.u.cadence.internal.worker.Poller - start(): Poller{options=PollerOptions{maximumPollRateIntervalMilliseconds=1000, maximumPollRatePerSecond=0.0, pollBackoffCoefficient=2.0, pollBackoffInitialInterval=PT0.2S, pollBackoffMaximumInterval=PT20S, pollThreadCount=1, pollThreadNamePrefix='null'}, identity=81b8d0ac-ff89-47e8-b842-3dd26337feea}
``` 
No Hello printed. It is expected as a worker is just a workflow code host. The workflow has to be started to execute. Let's use Cadence CLI to start the workflow:
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
Let's look at the workflow execution history:
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
Even for a such trivial workflow the history gives a lot of useful information. For complex workflows it is really useful tools for production and development troubleshooting.
History can be automatically archived to a long term blob store (for example S3) upon workflow completion for compliance, analytical and troubleshooting purposes.
### Workflow ID Uniqueness
Before proceeding to a more complex workflow implementation let's look at the workflow ID semantic. 
When starting a workflow without providing an ID the client generates one in the form of an UUID. In most real life scenarios it is not a desired behavior. 
The business ID should be used instead. Let's specify the ID when starting a workflow:
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
Oops, Cadence doesn't let to create workflow with the same ID. But there are use cases when it is desired. For example there is a need to reexecute the workflow for whatever reason.
This is achieved by specifying a special flag _Workflow ID Reuse Policy_. The value of 1 means `AllowDuplicate`:
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
Now it might be clear why every workflow has two IDs: Workflow ID and Run ID. As the Workflow ID can be reused the Run ID uniquely identifies a particular run of a workflow. Run ID is 
system generated and cannot be controlled by client code.

Note that ID Reuse Policy applies only when previous run of a workflow is completed. 
Under no circumstances Cadence allows more than one instance of open workflow with the same ID. 

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
   --workflowidreusepolicy value, --wrp value  Optional input to configure if the same workflow ID is allow to use for new workflow execution. Available options: 0: AllowDuplicateFailedOnly, 1: AllowDuplicate, 2: RejectDuplicate (default: 0)
   --input value, -i value                     Optional input for the workflow, in JSON format. If there are multiple parameters, concatenate them and separate by space.
   --input_file value, --if value              Optional input for the workflow from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. Input from file will be overwrite by input from command line
   --memo_key value                            Optional key of memo. If there are multiple keys, concatenate them and separate by space
   --memo value                                Optional info that can be showed when list workflow, in JSON format. If there are multiple JSON, concatenate them and separate by space. The order must be same as memo_key
   --memo_file value                           Optional info that can be listed in list workflow, from JSON format file. If there are multiple JSON, concatenate them and separate by space or newline. The order must be same as memo_key
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
  }
```
The workflow interface now has a new method annotated with @SignalMethod. It is a callback method that is invoked
every time a new signal of "HelloWorld::updateGreeting" is delivered to a workflow. The workflow interface can have only
one @WorkflowMethod which is a _main_ function of the workflow and as many signal methods as needed.

The updated workflow implementation demonstrates a few important Cadence concepts. The first is that workflow is stateful and can 
have fields of any complex type. Another one is `Workflow.await` function that blocks until the function it receives as a parameter evaluates to true.
The condition is going to be evaluated only on workflow state changes, so it is not a busy wait in traditional sense.   
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
Try sending the same signal with the same input again. Note that output doesn't change. It happens because await condition
doesn't unblock when it sees the same value. By a new greeting unblocks it:
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
Now shutdown the worker and send the same signal again:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloSignal" --name "HelloWorld::updateGreeting" --input \"Welcome\"
Signal workflow succeeded.
```
Note that sending signals as well as starting workflows does not need a worker running. The requests are queued inside the Cadence service. 

Now bring the worker back. Note that it doesn't log anything besides the standard startup messages.
It happens because it ignores the queued signal that contains the same input as the current value of greeting. 
Note that the restart of the worker didn't affect the workflow execution. It is still blocked on the same line of code as before the failure.
This is the most important feature of Cadence. The workflow code doesn't need to deal with worker failures at all. It state is fully recovered to its current state that includes all the
local variables and threads.

Let's look at which line the workflow is blocked:
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
Let's complete the workflow by sending a signal with "Bye" greeting:
 
```text
16:58:22.962 [workflow-root] INFO  c.u.c.samples.hello.GettingStarted - 4: Bye World!
```
Note that value of count variable was not lost during the restart. 

Also note that while a single worker instance is used for this
walk through, any real production deployment has multiple worker instances running. So any worker failure or restart does not delay any
workflow execution as it is just migrated to any other available worker.
## Query
So far we learned that the workflow code is fault tolerant and can update its state in reaction to external events in form of signals.
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
It also is not allowed to block its thread in any way. It is usually just returns a value derived from the fields of the workflow object.
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
Now let's query the workflow using CLI:
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow query --workflow_id "HelloQuery" --query_type "HelloWorld::getCount"
Query result as JSON:
3
```
One limitation of the query is that it requires a worker process running as it is executing callback code. 
An interesting feature of the query is that it works for completed workflows as well. Let's complete the workflow by sending "Bye" and query it.
```bash
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow signal --workflow_id "HelloQuery" --name "HelloWorld::updateGreeting" --input \"Bye\"
Signal workflow succeeded.
cadence: docker run --network=host --rm ubercadence/cli:master --do test-domain workflow query --workflow_id "HelloQuery" --query_type "HelloWorld::getCount"
Query result as JSON:
4
```
Query method can accept parameters. It might be useful if only part of the workflow state should be returned.
## Activities