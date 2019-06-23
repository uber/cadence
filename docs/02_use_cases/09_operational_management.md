# Operational Management

Imagine that you have to create a self operating database similar to Amazon RDS. Cadence is used in multiple projects
that automate managing and automatic recovery of various products like MySQL, Elastic Search and Cassandra.

Such systems usually is a mixture of different use cases. They need to monitor status of resources using polling. They have to execute orchestration API calls to
administrative interfaces of a database. They have to provision new hardware or docker instances if necessary. They need to push configuration updates and perform other actions like backups periodically.

