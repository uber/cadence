# How Cadence works

The Cadence hosted service brokers and persists events generated during workflow execution. Worker
nodes owned and operated by customers execute the coordination and task logic. To facilitate the
implementation of worker nodes, Cadence provides a client-side library for Go and Java. See Figure 1
where:

* FE: Front end
* HS: History service
* MS: Matching service

![Cadence overview diagram]({{ '/assets/overview.png' | relative_url }})

   **Figure 1**

In Cadence, you can code the logical flow of events separately as a workflow and code business logic
as activities. The workflow identifies the activities and sequences them, while an activity executes
the logic.

The video below explores the Cadence architecture in more depth:

<iframe width="560" height="315" src="https://www.youtube.com/embed/5M5eiNBUf4Q" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
