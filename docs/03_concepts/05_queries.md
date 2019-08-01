# Synchronous Query

Workflow code is stateful with the Cadence framework preserving it over various software and hardware failures. The state is constantly mutated during workflow execution. Cadence provides synchronous query capability that allows for the workflow expose any internal data. From the workflow implmenter point of view the query is exposed as a synchronous callback that is invoked by external entities. Multiple such callbacks can be provided per workflow type exposing different information to different external systems.

To execute a query an external client must provide _domian, workflowID, query name_ and optional _query arguments_.

Query callbacks must be read-only and cannot mutate the workflow state in any way.

## Stack Trace Query

The Cadence client libraries expose some predefined queries out of the box. Currently the only supported built-in query is _stack trace_. This query returns stacks of all workflow owned threads. This is a great way to troubleshoot any workflow in production.