# Tracing and Context Propagation

## Tracing

The Go client provides distributed tracing support through [opentracing](https://opentracing.io/). Tracing can be
configured by providing a `opentracing.Tracer` implementation in `Client` and `Worker` instantiation. Tracing allows
one to view the call graph of a workflow alongwith its activities, child workflows etc. For more details on how to
configure and leverage tracing please look at the [opentracing documentation](https://opentracing.io/docs/getting-started/).
The opentracing support has been validated using [Jaeger](https://www.jaegertracing.io/), but other implementations
mentioned [here](https://opentracing.io/docs/supported-tracers/) should work. Tracing support utilizes generic context
propagation support provided by the client.

## Context Propagation

We provide a standard way to propagate custom context across a workflow. For a sample, the Go client implements a
[tracing context propagator](https://github.com/uber-go/cadence-client/blob/master/internal/tracer.go).

### Server-Side Headers Support

On the server side Cadence provides a mechanism to propagate what it calls headers across different workflow
transitions.

```go
struct Header {
  10: optional map<string, binary> fields
}
```

The client leverages this to pass around selected context information. `HeaderReader` and `HeaderWriter` are interfaces
that allow reading and writing to the cadence server headers. The client already provides [implementations](https://github.com/uber-go/cadence-client/blob/master/internal/headers.go)
for these. `HeaderWriter` sets a field in the header. Headers is a map, so setting a value for the the same key
multiple times will overwrite the previous values. `HeaderReader` iterates through the headers map and runs the
provided handler function on each key value pair, allowing the user to deal with the fields one is interested in.

```go
type HeaderWriter interface {
  Set(string, []byte)
}

type HeaderReader interface {
  ForEachKey(handler func(string, []byte) error) error
}
```

### Context Propagators

Context propagators require implementing the following 4 methods to propagate selected context across a workflow.
`Inject` is meant to pick out the context keys of interest from a Go `context.Context` object and write that into the
headers using the `HeaderWriter` interface. `InjectFromWorkflow` is the same, but operates on a `workflow.Context`
object. `Extract` and `ExtractToWorkflow` read the headers and place the information of interest back into the
`context.Context` and `workflow.Context` respectively. The [tracing context propagator]([tracing context propagator](https://github.com/uber-go/cadence-client/blob/master/internal/tracer.go)
shows a sample implementation of context propagation.

```go
type ContextPropagator interface {
  Inject(context.Context, HeaderWriter) error

  Extract(context.Context, HeaderReader) (context.Context, error)

  InjectFromWorkflow(Context, HeaderWriter) error

  ExtractToWorkflow(Context, HeaderReader) (Context, error)
}
```
