ALTER TABLE open_executions ADD parent_run_id timestamp;
ALTER TABLE closed_executions ADD parent_run_id timestamp;
ALTER TABLE closed_executions_v2 ADD parent_run_id timestamp;