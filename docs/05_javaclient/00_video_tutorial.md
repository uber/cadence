# Video Tutorials

## Installing Cadence Service and UI on Mac

{% include youtubePlayer.html id="aLyRyNe5Ls0" %}

Commands executed during the tutorial:

```bash
docker-compose up

docker run --rm ubercadence/cli:master --address host.docker.internal:7933 --domain samples-domain domain register

docker run --rm ubercadence/cli:master --address host.docker.internal:7933 --domain samples-domain domain describe

alias cadence="docker run --rm ubercadence/cli:master --address host.docker.internal:7933"

cadence --domain samples-domain domain desc

cadence help

cadence workflow help

cadence --domain samples-domain workflow list

cadence --domain samples-domain workflow help start

cadence --domain samples-domain workflow start -wt test -tl test -et 300

cadence --domain samples-domain workflow list -op

cadence --domain samples-domain workflow terminate -wid <workflowID>

```

## Java Hello World Workflow Implementation

{% include youtubePlayer.html id="5mBLspVKOAI" %}

Source code:

```java
public interface HelloWorkflow {

@WorkflowMethod(executionStartToCloseTimeoutSeconds = 300)
String getGreeting(String name);

}
```
```java
package com.tutorial;

import com.uber.cadence.workflow.Workflow;

import java.time.Duration;

public class HelloWorkflowImpl implements HelloWorkflow {
    @Override
    public String getGreeting(String name) {
        Workflow.sleep(Duration.ofMinutes(1));
        return "Hello " + name + "!";
    }

}
```
```java
package com.tutorial;

import com.uber.cadence.worker.Worker;

public class Main {

    public static void main(String[] args) {
        Worker.Factory f = new Worker.Factory("samples-domain");
        Worker w = f.newWorker("hello");
        w.registerWorkflowImplementationTypes(HelloWorkflowImpl.class);
        f.start();
    }
}
```
Commands:
```bash
cadence -do samples-domain workflow start --et 300 --tl hello --wt HelloWorkflow::getGreeting --input \"World\"
```

