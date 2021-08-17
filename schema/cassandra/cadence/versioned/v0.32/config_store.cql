CREATE TABLE cluster_config (
  row_type int,
  version int,
  timestamp timestamp,
  values blob,
  encoding text,
PRIMARY KEY (row_type, version)
) WITH CLUSTERING ORDER BY (version DESC);