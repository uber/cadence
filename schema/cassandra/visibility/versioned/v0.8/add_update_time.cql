ALTER TABLE open_executions ADD update_time timestamp;
ALTER TABLE closed_executions ADD update_time timestamp;
ALTER TABLE closed_executions_v2 ADD update_time timestamp;
