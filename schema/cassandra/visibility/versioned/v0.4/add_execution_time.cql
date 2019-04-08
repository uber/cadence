ALTER TABLE open_executions ADD execution_time timestamp;
ALTER TABLE closed_executions ADD execution_time timestamp;
ALTER TABLE closed_executions_v2 ADD execution_time timestamp;

CREATE INDEX open_by_execution_time ON open_executions (execution_time);
CREATE INDEX closed_by_execution_time ON closed_executions (execution_time);
CREATE INDEX closed_by_execution_time_v2 ON closed_executions_v2 (execution_time);