---
codecontent: workerscaffoldgo04
weight: 35
categories: [tour]
---

# Building the Cadence Workflow Service Client

Cadence uses an RPC protocol called [TChannel](https://github.com/uber/tchannel). You only need to 
provide a name for the channel and let YARPC own and manage it. 

```go
ch, err := tchannel.NewChannelTransport(tchannel.ServiceName("cadence-client"))
```

The service name provided when creating the channel ("cadence-client") is for the worker, not for 
the Cadence server. 

The Cadence service is YARPC application using TChannel. In order for your worker to talk to it, 
you'll need to create a YARPC dispatcher with a TChannel "outbound".

```go
dispatcher := yarpc.NewDispatcher(yarpc.Config{
    Name: "cadence-client",
    Outbounds: yarpc.Outbounds{
        "cadence-frontend": {Unary: ch.NewSingleOutbound("127.0.0.1:7933")},
    },
})
```

"cadence-frontend" refers to the Cadence service the worker will make requests to. The outbound is 
hard-coded to use the local Cadence service running on Docker. You will want to make this configurable 
to support using a remote Cadence service. 

After starting the YARPC dispatcher, you can create a new Workflow Service Client and let it manage 
the dispatcher.

```go
dispatcher.Start()
h.Service = workflowserviceclient.New(dispatcher.ClientConfig("cadence-frontend"))
```
