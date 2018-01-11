CREATE TYPE domain_config (
  retention   int,
  emit_metric boolean
);

CREATE TYPE cluster_replication_config (
  cluster_name text,
);

ALTER TABLE domains_by_name ADD replication_config frozen<domain_replication_config>;
ALTER TABLE domains_by_name ADD version bigint;