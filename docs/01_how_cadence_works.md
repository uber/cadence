---
redirect_from: 
   - /docs/how_cadence_works
   - docs/how_cadence_works
---

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
