---
layout: default
title: Task lists
permalink: /docs/concepts/task-lists
---

# Task lists

When a :workflow: invokes an :activity:, it sends the ```ScheduleActivityTask``` :decision: to the
Cadence service. As a result, the service updates the :workflow: state and dispatches
an :activity_task: to a :worker: that implements the :activity:.
Instead of calling the :worker: directly, an intermediate queue is used. So the service adds an _:activity_task:_ to this
queue and a :worker: receives the :task: using a long poll request.
Cadence calls this queue used to dispatch :activity_task:activity_tasks: an *:activity_task_list:*.

Similarly, when a :workflow: needs to handle an external :event:, a :decision_task: is created.
A :decision_task_list: is used to deliver it to the :workflow_worker: (also called _decider_).

While Cadence :task_list:task_lists: are queues, they have some differences from commonly used queuing technologies.
The main one is that they do not require explicit registration and are created on demand. The number of :task_list:task_lists:
is not limited. A common use case is to have a :task_list: per :worker: process and use it to deliver :activity_task:activity_tasks:
to the process. Another use case is to have a :task_list: per pool of :worker:workers:.

There are multiple advantages of using a :task_list: to deliver :task:tasks: instead of invoking an :activity_worker: through a synchronous RPC:

* :worker:Worker: doesn't need to have any open ports, which is more secure.
* :worker:Worker: doesn't need to advertise itself through DNS or any other network discovery mechanism.
* When all :worker:workers: are down, messages are persisted in a :task_list: waiting for the :worker:workers: to recover.
* A :worker: polls for a message only when it has spare capacity, so it never gets overloaded.
* Automatic load balancing across a large number of :worker:workers:.
* :task_list:Task_lists: support server side throttling. This allows you to limit the :task: dispatch rate to the pool of :worker:workers: and still supports adding a :task: with a higher rate when spikes happen.
* :task_list:Task_lists: can be used to route a request to specific pools of :worker:workers: or even a specific process.
