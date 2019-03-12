# Workers

A worker or *worker service* is a service that hosts the workflow and activity implementations. The worker polls the *Cadence service* for tasks, performs those tasks and communicates task execution results back to the *Cadence service*. Worker services are developed, deployed and operated by Cadence customers.

You can run a Cadence worker in a new or an exiting service. Use the framework APIs to start the Cadence worker and link in all activity and workflow implementations that you require the service to execute.

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
var CadenceService = "CadenceServiceFrontend"

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
	// TaskListName - identifies set of client workflows, activities and workers.
	// it could be your group or client or application name.
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, map[string]string{}),
	}

	worker := worker.NewWorker(
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
