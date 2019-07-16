# Go Client

## Overview

Go client attemps to follow Go language conventions. The conversion of a go program to the fault-oblivious workflow function is expected to be pretty mechanical.

Cadence requires determinism of the workflow code. It supports deterministic execution of the multithreaded code and constructs like ``select`` that are non deterministic by Go design. The Cadence solution is to provide corresponding constructs in form of interfaces that have similar capability but support deterministic execution.

For example instead of native Go channels workflow code must use workflow.Channel interface. Instead of ``select`` the Selector interface must be used.

See [Creating Workflows](02_create_workflows#]) for more info.

## Links

- GitHub project: [https://github.com/uber-go/cadence-client](https://github.com/uber-go/cadence-client)
- Samples: [https://github.com/samarabbas/cadence-samples](https://github.com/samarabbas/cadence-samples)
- GoDoc documentation: [https://godoc.org/go.uber.org/cadence](https://godoc.org/go.uber.org/cadence)
