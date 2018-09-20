-- for workflow retry
ALTER TYPE workflow_execution ADD attempt              int;
ALTER TYPE workflow_execution ADD has_retry_policy     boolean;
ALTER TYPE workflow_execution ADD init_interval        int;
ALTER TYPE workflow_execution ADD backoff_coefficient  double;
ALTER TYPE workflow_execution ADD max_interval         int;
ALTER TYPE workflow_execution ADD expiration_time      timestamp;
ALTER TYPE workflow_execution ADD max_attempts         int;
ALTER TYPE workflow_execution ADD non_retriable_errors list<text>;
-- for blob headers
ALTER TABLE events ADD data_headers map<text, text>;
ALTER TYPE workflow_execution ADD completion_event_blob_headers  map<text,text>;
ALTER TYPE workflow_execution ADD execution_context_blob_headers  map<text,text>;
ALTER TYPE activity_info ADD blob_headers  map<text,text>;
ALTER TYPE child_execution_info ADD blob_headers  map<text,text>;
ALTER TYPE signal_info ADD blob_headers  map<text,text> ;
