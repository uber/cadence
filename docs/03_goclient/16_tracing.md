# Tracing and Context Propagation

## Tracing

The Go client provides distributed tracing support through [opentracing](https://opentracing.io/). Tracing can be
configured by providing an [opentracing.Tracer](https://godoc.org/github.com/opentracing/opentracing-go#Tracer)
implementation in [ClientOptions](https://godoc.org/go.uber.org/cadence/internal#ClientOptions)
and [WorkerOptions](https://godoc.org/go.uber.org/cadence/internal#WorkerOptions) during client and worker instantiation
respectively. Tracing allows
you to view the call graph of a workflow along with its activities, child workflows etc. For more details on how to
configure and leverage tracing please look at the [opentracing documentation](https://opentracing.io/docs/getting-started/).
The opentracing support has been validated using [Jaeger](https://www.jaegertracing.io/), but other implementations
mentioned [here](https://opentracing.io/docs/supported-tracers/) should also work. Tracing support utilizes generic context
propagation support provided by the client.

## Context Propagation

We provide a standard way to propagate custom context across a workflow. The
[ClientOptions](https://godoc.org/go.uber.org/cadence/internal#ClientOptions) and [WorkerOptions](https://godoc.org/go.uber.org/cadence/internal#WorkerOptions)
allow configuring a context propagator. Context propagator extracts and passes on information present in the `context.Context`
and `workflow.Context` objects across workflow. Once a context propagator is configured you should be able to access the required values
in the context objects as you would normally do in Go.
For a sample, the Go client implements a [tracing context propagator](https://github.com/uber-go/cadence-client/blob/master/internal/tracer.go).

### Server-Side Headers Support

On the server side Cadence provides a mechanism to propagate what it calls headers across different workflow
transitions.

```go
struct Header {
  10: optional map<string, binary> fields
}
```

The client leverages this to pass around selected context information. [HeaderReader](https://godoc.org/go.uber.org/cadence/internal#HeaderReader)
and [HeaderWriter](https://godoc.org/go.uber.org/cadence/internal#HeaderWriter) are interfaces
that allow reading and writing to the Cadence server headers. The client already provides [implementations](https://github.com/uber-go/cadence-client/blob/master/internal/headers.go)
for these. `HeaderWriter` sets a field in the header. Headers is a map, so setting a value for the the same key
multiple times will overwrite the previous values. `HeaderReader` iterates through the headers map and runs the
provided handler function on each key value pair, allowing you to deal with the fields you are interested in.

```go
type HeaderWriter interface {
  Set(string, []byte)
}

type HeaderReader interface {
  ForEachKey(handler func(string, []byte) error) error
}
```

### Context Propagators

Context propagators require implementing the following 4 methods to propagate selected context across a workflow:

- `Inject` is meant to pick out the context keys of interest from a Go [context.Context](https://golang.org/pkg/context/#Context) object and write that into the
headers using the [HeaderWriter](https://godoc.org/go.uber.org/cadence/internal#HeaderWriter) interface
- `InjectFromWorkflow` is the same as above, but operates on a [workflow.Context](https://godoc.org/go.uber.org/cadence/internal#Context) object
- `Extract` read the headers and place the information of interest back into the [context.Context](https://golang.org/pkg/context/#Context) object
- `ExtractToWorkflow` is the same as above, but operates on a [workflow.Context](https://godoc.org/go.uber.org/cadence/internal#Context) object

The [tracing context propagator]([tracing context propagator](https://github.com/uber-go/cadence-client/blob/master/internal/tracer.go)
shows a sample implementation of context propagation.

```go
type ContextPropagator interface {
  Inject(context.Context, HeaderWriter) error

  Extract(context.Context, HeaderReader) (context.Context, error)

  InjectFromWorkflow(Context, HeaderWriter) error

  ExtractToWorkflow(Context, HeaderReader) (Context, error)
}
```

## Q&A

### Is there a complete example?

We will have a sample available soon.

### Can I configure multiple context propagators?

Yes, it is recommended to configure multiple context propagators with each propagator meant to propagate a particular type of context.
