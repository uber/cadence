# DSL Workflows

Cadence supports implementing business logic directly in programming languages like Java and Go. But there are cases when
using a domain-specific language is more appropriate. Or there might be a legacy system that uses some form of DSL for process definition but it is not operationally stable and scalable. This also applies to more recent systems like Apache Airflow, various BPMN engines and AWS Step Functions.

An application that interprets the DSL definition can be written using the Cadence SDK. It automatically becomes highly fault tolerant, scalable, and durable when running on Cadence. Cadence has been used to deprecate several Uber internal DSL engines. The customers continue to use existing process definitions, but Cadence is used as an execution engine.

There are multiple benefits of unifying all company workflow engines on top of Cadence. The most obvious one is that
it is more efficient to support a single product instead of many. It is also difficult to beat the scalability and stability of
Cadence which each of the integrations it comes with. Additionally, the ability to share activities across "engines"
might be a huge benefit in some cases.
