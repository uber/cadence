---
codecontent: workerscaffoldgo06
weight: 45
categories: [tour]
---

# Starting the Worker

Now that the Workflow Service Client is setup, the next step is to start the worker. Be sure 
to [start a local server]({{ '/tour/start_local_server' | relative_url }}) from another terminal 
and [create the sample domain]({{ '/tour/create_sample_domain' | relative_url }}) before 
starting the worker.

A worker can only monitor one domain and one 
[task list]({{ '/docs/07_glossary#task-list' | relative_url }}). 