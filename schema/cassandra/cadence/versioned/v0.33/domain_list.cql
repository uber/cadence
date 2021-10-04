CREATE TABLE executions (
  shard_id                       int,
  type                           int, -- enum RowType { Shard, Execution, TransferTask, TimerTask, ReplicationTask, CrossClusterTask}
  domain_id                      uuid,
  workflow_id                    text,
  run_id                         uuid,
  current_run_id                 uuid,
  visibility_ts                  timestamp, -- unique identifier for timer tasks for an execution
  task_id                        bigint, -- unique identifier for transfer and timer tasks for an execution
  shard                          frozen<shard>,
  execution                      frozen<workflow_execution>,
  transfer                       frozen<transfer_task>,
  cross_cluster                  frozen<transfer_task>, -- reuse the transfer_task type
  replication                    frozen<replication_task>,
  timer                          frozen<timer_task>,
  next_event_id                  bigint,  -- This is needed to make conditional updates on session history
  range_id                       bigint, -- Increasing sequence identifier for transfer queue, checkpointed into shard info
  activity_map                   map<bigint, frozen<activity_info>>,
  timer_map                      map<text, frozen<timer_info>>,
  child_executions_map           map<bigint, frozen<child_execution_info>>,
  request_cancel_map             map<bigint, frozen<request_cancel_info>>,
  signal_map                     map<bigint, frozen<signal_info>>,
  signal_requested               set<uuid>,
  buffered_events_list           list<frozen<serialized_event_batch>>,
  replication_state              frozen<replication_state>, -- Replication information part of mutable state
  buffered_replication_tasks_map map<bigint, frozen<buffered_replication_task_info>>,
  workflow_last_write_version    bigint,
  workflow_state                 int,
  version_histories              blob, -- the metadata of history branching
  version_histories_encoding     text,
  checksum                       frozen<checksum>,
  domain_list                    list<text>,
  PRIMARY KEY  (shard_id, type, domain_id, workflow_id, run_id, visibility_ts, task_id)
) WITH COMPACTION = {
    'class': 'org.apache.cassandra.db.compaction.LeveledCompactionStrategy'
  };