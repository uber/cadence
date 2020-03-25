CREATE TYPE domain_failover_config (failover_start_time timestamp, failover_timeout int);
ALTER TABLE domains_by_name_v2 ADD failover_config frozen<domain_failover_config>;