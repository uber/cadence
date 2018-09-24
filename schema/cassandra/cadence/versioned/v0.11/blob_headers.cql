-- for blob headers
ALTER TABLE events ADD data_headers map<text, text>;
ALTER TYPE workflow_execution ADD completion_event_data_headers  map<text,text>;
ALTER TYPE workflow_execution ADD execution_context_data_headers  map<text,text>;
ALTER TYPE activity_info ADD event_data_headers  map<text,text>;
ALTER TYPE activity_info ADD detail_data_headers  map<text,text>;
ALTER TYPE child_execution_info ADD event_data_headers  map<text,text>;
ALTER TYPE signal_info ADD data_headers  map<text,text> ;
ALTER TYPE serialized_event_batch ADD data_headers  map<text, text>;
