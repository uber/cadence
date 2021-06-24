ALTER TYPE shard ADD cross_cluster_processing_queue_states blob;
ALTER TYPE shard ADD cross_cluster_processing_queue_states_encoding text;

ALTER TABLE executions ADD cross_cluster frozen<transfer_task>;