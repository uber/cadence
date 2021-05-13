ALTER TABLE open_executions ADD is_cron boolean;
ALTER TABLE closed_executions ADD is_cron boolean;
ALTER TABLE closed_executions_v2 ADD is_cron boolean;
