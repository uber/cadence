ALTER TYPE workflow_execution ADD decision_attempt bigint;
ALTER TYPE workflow_execution ADD failed_decision_started_id bigint;

ALTER TYPE transfer_task ADD schedule_attempt bigint;
ALTER TYPE task ADD schedule_attempt bigint;