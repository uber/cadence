# Flow of tasks

Below diagram explains (to certain level of detail) how a workflow is processed by Cadence workers and backend components.

[comment]: <> (To visualize mermaid flowchart below, install Mermaid plugin for your IDE. Works in github out of the box)

```mermaid
flowchart TD
    User((Domain X Admin)) -->|"0. Deploy worker with config: domain, tasklist(s)"| Worker[[\nDomain X Worker\n\n]]
    ExternalInitiator(("Domain X\nAnother Service/CLI")) -->|"1. Start/Signal Workflow"| Frontend
    Worker -->|2. Poll tasks for domain: X, tasklist: tl1| Frontend[[\nCadence Frontend\n\n]]
    Frontend -->|3. Forward Poll request| Matching[[\nCadence Matching\n\n]]
    Matching -->|4. Poll| TaskListManager(Tasklist Manager for tl1)
    TaskListManager -->|5. Long Poll| WaitUntilTaskFound
    WaitUntilTaskFound --> |6. Return task to worker| Worker
    Worker -->|7. Generate decisions\n& respond| Frontend
    Frontend -->|8. Respond decisions| History[[\nCadence History\n\n]]
    History -->|9. Mutable state update| ExecutionsTable[(Executions Table)]
    History -->|10.a. Notify| HistoryQueue(History Queues - per shard)
    HistoryQueue -->|Periodic task fetch| ExecutionsTable
    HistoryQueue -->|11.b. Execute transfer task| TransferTaskExecutor(Transfer Task Executor)
    HistoryQueue -->|11.a. Execute timer task| TimerTaskExecutor(Timer Task Executor)
    HistoryQueue -->|Periodic per-shard offset save| ShardsTable[(Shards Table)]
    TransferTaskExecutor -->|12.b. Add task| Matching
    TimerTaskExecutor -->|12.a. Add task| Matching
    Matching -->|13. Add task| TaskListManager
    TaskListManager -->|14. Add task| SyncMatch{Check sync match}
    SyncMatch -->|15.b. Found poller - Do sync match| SyncMatched
    SyncMatch -->|15.a. No pollers - Save task| TasksTable[(Tasks Table)]
    TaskListManager -->|Periodic task fetch| TasksTable

```
