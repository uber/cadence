# Use Cases

As Cadence developers we face a hard non technical problem. It is to how position and describe the Cadence platform.

We call it _worklfow_. But when most people hear the word workflow they think about low-code and UIs. While these might be useful for non technical users, they frequently bring more pain than value to software engineers. Most UIs and low-code DSLs are awesome for demo applications, but any diagram with 100+ elements or a few thousand lines JSON DSL is completely unpractical. So positioning Cadence as a workflow is not ideal as it turns off developers that would enjoy its code only approach.

We call it _orchestrator_. But this term is used for multiple purposes and turns off customers that want to implement business process automation solutions.

We call it _durable function_ platform. It is technically a correct name. But most developers outside of Microsoft ecosystem never heard of durable functions.

We believe that problem in naming comes from the fact that Cadence is indeed a **new way to write distributed applications**. It is so generic that it can be applied to practically any use case that goes beyond a single request reply. It can be used to build applications that are in tranditional areas of workflow or orchestration platforms. But it also huge _developer productivity_ boost for multiple use cases that traditionally rely on databases or/and task queues.

This section represents far from complete list of use case types where Cadence is a good fit. All of them have been used by real production services at Uber and outside.

Don't treat the list as exclusive. It is common to employ mutliple of those types in a single application. For example an operational management use case would need periodic execution, service orchestration, polling, event driven as well as interactive parts.
