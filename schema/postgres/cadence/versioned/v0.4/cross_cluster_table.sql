CREATE TABLE cross_cluster_tasks(
  target_cluster VARCHAR(255) NOT NULL,
  shard_id INTEGER NOT NULL,
  task_id BIGINT NOT NULL,
  --
  data BYTEA NOT NULL,
  data_encoding VARCHAR(16) NOT NULL,
  PRIMARY KEY (target_cluster, shard_id, task_id)
);