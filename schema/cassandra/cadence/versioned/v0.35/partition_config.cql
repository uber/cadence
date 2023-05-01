ALTER TYPE workflow_execution ADD partition_config map<text, text>;
ALTER TYPE task ADD partition_config map<text, text>;
