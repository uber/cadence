ALTER TYPE workflow_execution ADD decision_schedule_time timestamp;
ALTER TYPE child_execution_info ADD initiated_time timestamp;
ALTER TYPE request_cancel_info ADD initiated_time timestamp;
ALTER TYPE signal_info ADD initiated_time timestamp;