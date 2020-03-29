ALTER TYPE shard ADD replication_dlq_ack_level map<text, bigint>;
ALTER TABLE queue ALTER message_id TYPE bigint;