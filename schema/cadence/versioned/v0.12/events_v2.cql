CREATE TYPE branch_range (
  branch_id   uuid,
  forked_node_id bigint,
);

CREATE TABLE events_v2 (
  tree_id           uuid,
  branch_id         uuid,
  row_type          bigint, -- 0: branch row, 1: node row

  ancestors         list<frozen<branch_range>>, -- Only for branch row.
  deleted           boolean, -- Only for branch row

  node_id           bigint, -- Only for node row: 0: root node, others: regular node
  txn_id            bigint, -- Only for node row: for optimistic locking. For root node it is a tree operation transactionID, for regular node it is for override update when 2PC update fails.

  data                blob, -- Only for node row: Batch of workflow execution history events as a blob
  data_encoding       text, -- Only for node row: Protocol used for history serialization
  data_version        int,  -- Only for node row: history blob version
PRIMARY KEY ((tree_id), row_type, branch_id, node_id )
) WITH COMPACTION = {
  'class': 'org.apache.cassandra.db.compaction.LeveledCompactionStrategy'
};