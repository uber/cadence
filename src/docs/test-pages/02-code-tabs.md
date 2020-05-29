---
layout: default
title: Code tabs
permalink: /docs/test-page/code-tabs
---

# Code tabs

This is an example of code tabs

<CodeSwitcher :languages="{go:'Go',java:'Java'}">
<template v-slot:go>

``` go
package sample

import (
    "time"

    "go.uber.org/cadence/workflow"
)

func init() {
    workflow.Register(SimpleWorkflow)
}
```

</template>
<template v-slot:java>

```java
public class FileProcessingWorkflowImpl implements FileProcessingWorkflow {

    private final FileProcessingActivities activities;

    public FileProcessingWorkflowImpl() {
        this.activities = Workflow.newActivityStub(FileProcessingActivities.class);
    }

    ...
}
```

</template>
</CodeSwitcher>

And another example also updates to selected language

<CodeSwitcher :languages="{go:'Go',java:'Java'}">
<template v-slot:go>

```go
package main

import (

	t "go.uber.org/cadence/.gen/go/cadence"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/worker"

	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/tchannel"
)

var HostPort = "127.0.0.1:7933"
var Domain = "SimpleDomain"
var TaskListName = "SimpleWorker"
var ClientName = "SimpleWorker"
var CadenceService = "cadence-frontend"

func main() {
	startWorker(buildLogger(), buildCadenceClient())
}

func buildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger")
	}

	return logger
}

func buildCadenceClient() workflowserviceclient.Interface {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(ClientName))
	if err != nil {
		panic("Failed to setup tchannel")
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
			Name: ClientName,
			Outbounds: yarpc.Outbounds{
				CadenceService: {Unary: ch.NewSingleOutbound(HostPort)},
			},
		})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	return workflowserviceclient.New(dispatcher.ClientConfig(CadenceService))
}

func startWorker(logger *zap.Logger, service workflowserviceclient.Interface) {
	// TaskListName identifies set of client workflows, activities, and workers.
	// It could be your group or client or application name.
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, map[string]string{}),
	}

	worker := worker.New(
		service,
		Domain,
		TaskListName,
		workerOptions)
	err := worker.Start()
	if err != nil {
		panic("Failed to start worker")
	}

	logger.Info("Started Worker.", zap.String("worker", TaskListName))
}
```

</template>
<template v-slot:java>

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

</template>
</CodeSwitcher>
