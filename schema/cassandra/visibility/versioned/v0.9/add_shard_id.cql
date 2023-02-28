ALTER TABLE open_executions ADD shard_id int;
ALTER TABLE closed_executions ADD shard_id int;
ALTER TABLE closed_executions_v2 ADD shard_id int;
