# Go Client

## Overview

Go client attempts to follow Go language conventions. The conversion of a Go program to the fault-oblivious workflow function is expected to be pretty mechanical.

Cadence requires determinism of the workflow code. It supports deterministic execution of the multithreaded code and constructs like `select` that are non-deterministic by Go design. The Cadence solution is to provide corresponding constructs in the form of interfaces that have similar capability but support deterministic execution.

For example, instead of native Go channels, workflow code must use the `workflow.Channel` interface. Instead of `select`, the Selector interface must be used.

For more information, see [Creating Workflows](02_create_workflows#]).

## Links

- GitHub project: [https://github.com/uber-go/cadence-client](https://github.com/uber-go/cadence-client)
- Samples: [https://github.com/samarabbas/cadence-samples](https://github.com/samarabbas/cadence-samples)
- GoDoc documentation: [https://godoc.org/go.uber.org/cadence](https://godoc.org/go.uber.org/cadence)
