ALTER TABLE open_executions ADD parent_workflow_id timestamp;
ALTER TABLE closed_executions ADD parent_workflow_id timestamp;
ALTER TABLE closed_executions_v2 ADD parent_workflow_id timestamp;