CREATE TABLE executions_visibility (
  domain_id            CHAR(64) NOT NULL,
  run_id               CHAR(64) NOT NULL,
  start_time           TIMESTAMP(3) NOT NULL,
  workflow_id          VARCHAR(255) NOT NULL,
  workflow_type_name   VARCHAR(255) NOT NULL,
  close_status         INT,  -- enum WorkflowExecutionCloseStatus {COMPLETED, FAILED, CANCELED, TERMINATED, CONTINUED_AS_NEW, TIMED_OUT}
  close_time           TIMESTAMP(3),
  history_length       BIGINT,

  PRIMARY KEY  (domain_id, run_id)
);

CREATE INDEX by_type_start_time ON executions_visibility (domain_id, close_status, start_time, run_id);
CREATE INDEX by_type_start_time ON executions_visibility (domain_id, workflow_type_name, close_status, start_time, run_id);
CREATE INDEX by_workflow_id_start_time ON executions_visibility (domain_id, workflow_id, close_status, start_time, run_id);
CREATE INDEX by_status_by_close_time ON executions_visibility (domain_id, close_status, start_time, run_id);
