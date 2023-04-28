ALTER TYPE workflow_execution ADD partition_config map<text, text>;
ALTER TYPE task ADD partition_config map<text, text>;

ALTER TYPE domain_config ADD isolation_groups blob;
ALTER TYPE domain_config ADD isolation_groups_encoding text;
