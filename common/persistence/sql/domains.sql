CREATE TABLE IF NOT EXISTS domains(
/* domain */
  id CHAR(36) PRIMARY KEY NOT NULL,
  name VARCHAR(255) UNIQUE NOT NULL,
  status INT NOT NULL,
  description VARCHAR(255) NOT NULL,
  owner_email VARCHAR(255) NOT NULL,
  data BLOB NOT NULL,
/* end domain */
  retention_days INT NOT NULL,
  emit_metric TINYINT(1) NOT NULL,
/* end domain_config */
  config_version BIGINT NOT NULL,
  notification_version BIGINT NOT NULL,
  failover_notification_version BIGINT NOT NULL,
  failover_version BIGINT NOT NULL,
  is_global_domain TINYINT(1) NOT NULL,
/* domain_replication_config */
  active_cluster_name VARCHAR(255) NOT NULL,
  clusters BLOB NOT NULL
/* end domain_replication_config */
) DEFAULT CHARACTER SET utf8 COLLATE utf8_unicode_ci;

CREATE TABLE IF NOT EXISTS domain_metadata (
  notification_version BIGINT NOT NULL
);

INSERT INTO domain_metadata (notification_version) VALUES (0);