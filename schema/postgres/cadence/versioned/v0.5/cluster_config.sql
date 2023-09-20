CREATE TABLE cluster_config (
  row_type INTEGER NOT NULL,
  version BIGINT NOT NULL,
  --
  timestamp TIMESTAMP NOT NULL,
  data           BYTEA NOT NULL,
  data_encoding  VARCHAR(16) NOT NULL,
  PRIMARY KEY (row_type, version)
);
