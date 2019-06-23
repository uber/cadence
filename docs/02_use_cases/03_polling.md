# Polling

Polling is executing periodic action to check for state changes. Examples are pinging a host, calling REST API or listing S3 bucket for newly uploaded files.

Cadence support for long running activities and unlimited retries makes it a good fit.

Some real life use cases:

* Network, host and service monitoring
* Processing files uploaded to FTP or S3
* Polling external API for a specific resource to become available
