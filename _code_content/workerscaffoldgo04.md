---
name: workerscaffoldgo04
---

**common/sample_helper.go**
```go
func (h *SampleHelper) buildServiceClient() error {
	ch, err := tchannel.NewChannelTransport(
		tchannel.ServiceName("cadence-client"))
	if err != nil {
		h.Logger.Fatal("Failed to create transport channel", zap.Error(err))
	}

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: "cadence-client",
		Outbounds: yarpc.Outbounds{
			"cadence-frontend": {Unary: ch.NewSingleOutbound("127.0.0.1:7933")},
		},
	})

	if err := dispatcher.Start(); err != nil {
		h.Logger.Fatal("Failed to create outbound transport channel: %v", zap.Error(err))
	}

	h.Service = workflowserviceclient.New(dispatcher.ClientConfig("cadence-frontend"))

	return nil
}
```
