---
layout: default
title: Java hello world
permalink: /docs/tutorials/java-hello-world
---

# Java Hello World

## Workflow implementation

<figure class="video-container">
  <iframe
    src="https://www.youtube.com/embed/5mBLspVKOAI"
    frameborder="0"
    height="315"
    allowfullscreen
    width="560"></iframe>
</figure>

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

