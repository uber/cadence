ALTER TYPE workflow_execution ADD partition_config map<text, text>;
ALTER TYPE task ADD partition_config map<text, text>;

CREATE TYPE isolation_group_partition (
  name text,
  state int,
);

ALTER TYPE domain_config ADD isolation_groups frozen<list<isolation_group_partition>>;
