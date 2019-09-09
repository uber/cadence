# Use Cases

As Cadence developers, we face a difficult non-technical problem: How to position and describe the Cadence platform.

We call it _workflow_. But when most people hear the word "workflow" they think about [low-code](https://en.wikipedia.org/wiki/Low-code_development_platform) and UIs. While these might be useful for non technical users, they frequently bring more pain than value to software engineers. Most UIs and low-code [DSLs](https://en.wikipedia.org/wiki/Domain-specific_language) are awesome for "hello world" demo applications, but any diagram with 100+ elements or a few thousand lines of JSON DSL is completely impractical. So positioning Cadence as a workflow is not ideal as it turns away developers that would enjoy its code-only approach.

We call it _orchestrator_. But this term is [pretty narrow](https://en.wikipedia.org/wiki/Orchestration_(computing)) and turns away customers that want to implement business process automation solutions.

We call it _durable function_ platform. It is technically a correct term. But most developers outside of the Microsoft ecosystem have never heard of [Durable Functions](https://docs.microsoft.com/en-us/azure/azure-functions/durable/durable-functions-overview).

We believe that problem in naming comes from the fact that Cadence is indeed a **new way to write distributed applications**. It is generic enough that it can be applied to practically any use case that goes beyond a single request reply. It can be used to build applications that are in traditional areas of workflow or orchestration platforms. But it is also huge _developer productivity_ boost for multiple use cases that traditionally rely on databases and/or task queues.

This section represents a far from complete list of use cases where Cadence is a good fit. All of them have been used by real production services inside and outside of Uber.

Don't think of this list as exhaustive. It is common to employ multiple use types in a single application. For example, an operational management use case might need periodic execution, service orchestration, polling, event driven, as well as interactive parts.
