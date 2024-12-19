ALTER TYPE executions ADD created_time timestamp;
ALTER TYPE history_node ADD created_time timestamp;
ALTER TYPE history_tree ADD created_time timestamp;
ALTER TYPE domains ADD created_time timestamp;
ALTER TYPE domains_by_name_v2 ADD created_time timestamp;
ALTER TYPE queue ADD created_time timestamp;
ALTER TYPE queue_metadata ADD created_time timestamp;
ALTER TYPE queue_metadata ADD last_updated_time timestamp;