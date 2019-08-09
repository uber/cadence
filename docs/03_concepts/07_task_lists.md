# Task Lists

When a workflow invokes an activity it sends the ```ScheduleActivityTask``` [decision](../04_glossary#decision) to the 
Cadence service. As a result the service updates the workflow state and dispatches 
an [activity task](../04_glossary#activity-task) to a worker that implements the activity. 
Instead of calling the worker directly an intermediate queue is used. So the service adds an _activity task_ to this 
queue and a worker receives the task using long poll request. 
Cadence calls this queue used to dispatch activity tasks *activity task list*.

Similarly when a workflow needs to handle an external event, a decision task is created. A so called *decision task list* is
used to deliver it to the workflow worker (also called _decider_).

While Cadence task lists are queues they have some differences from commonly used queuing technologies. 
The main one is that they do not require explicit registration and are created on demand. The number of task lists
is not limited. A common use case is to have a task list per worker process and use it to deliver activity tasks
to the process. Another use case is to have a task list per pool of workers.

There are multiple advantages of using a task list to deliver tasks instead of invoking an activity 
worker through a synchronous RPC:

* Worker doesn't need to have any open ports which is more secure.
* Worker doesn't need to advertise itself through DNS or any other network discovery mechanism.
* When all workers are down messages are persisted in a task list waiting for the workers to recover.
* Worker poll for a message only when it has a spare capacity. So it is never overloaded.
* Automatic load balancing across large number of workers
* Task lists support server side throttling. It allows to limit the task dispatch rate to the pool of workers and still support adding task with higher rate when spikes happen.
* Task lists can be used to route request to specific pools of workers or even specific process.
