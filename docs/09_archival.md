# Archival

Archival is a feature that automatically moves workflow histories from persistence to another location after the retention period. The purpose of archival is to be able to keep histories as long as needed while not overwhelming the persistence store. There are two reasons you may want to keep the histories after the retention period has past:
1. **Compliance:** For legal reasons histories may need to be stored for a long period of time.
2. **Debugging:** Old histories can still be accessed for debugging.

Archival is still in beta and there are three limits to its feature set:
1. **Only Histories:** Only histories are archived, visibility records are simply deleted after the retention period.
2. **RunID Required:** In order to access an archived history, both workflowID and runID are required.
3. **Best Effort:** There are cases in which a history can be deleted from persistence without being archived first. These cases are rare but are possible with the current state of archival.

Work is being prioritized on archival to eliminate these limitations.

## Concepts

- **Archiver:** Archiver is the component responsible for archiving and retrieving workflow histories.  Its interface is quite generic and supports different kinds of archival locations: local file system, S3, Kafka, etc. Check [this README](https://github.com/uber/cadence/blob/master/common/archiver/README.md) for how to add a new archiver implementation.
- **URI:** An URI is used to specify the archival location. Based on the scheme part of an URI, the corresponding archiver will be selected by the system to perform archival.

## Configuring Archival

Archival is controlled by both domain level config and cluster level config. 

### Cluster Archival Config

A Cadence cluster can be in one of three archival states:
  * **Disabled:** No archivals will occur and the archivers will be not initialized on startup.
  * **Paused:** This state is not yet implemented. Currently setting cluster to paused is the same as setting it to disabled.
  * **Enabled:** Archivals will occur.

Enabling the cluster for archival simply means histories are being archived. There is another config which controls whether histories can be accessed from archival. Both these configs have defaults defined in static yaml, and have dynamic config overwrites. Note, however, dynamic config will take effect only when archival is enabled in static yaml.

### Domain Archival Config

A domain includes two pieces of archival related config: 
  * **Status:** Either enabled or disabled. If a domain is in the disabled state no archivals will occur for that domain. 
  A domain can safely switch between statuses.
  * **URI:** The scheme and location where histories will be archived to. When a domain enables archival for the first time URI is set and can never be mutated. If URI is not specified when first enabling a domain for archival, a default URI from static config will be used.

## Running Locally

In order to run locally do the following:
1. `./cadence-server start`
2. `./cadence --do samples-domain domain register --gd false --history_archival_status enabled --retention 0`
3. Run the [helloworld cadence-sample](https://github.com/uber-common/cadence-samples) by following the README
4. Copy the workflowID and runID of the completed workflow from log output
5. `./cadence --do samples-domain wf show --wid <workflowID> --rid <runID>`

In step 2, we registered a new domain and enabled history archival feature for that domain. Since we didn't provide an archival URI when registering the new domain, the default URI specified in `config/development.yaml` is used. The default URI is `file:///tmp/cadence_archival/development`, so you can find the archived workflow history under the `/tmp/cadence_archival/development` directory. 

## FAQ

### How does archival interact with global domains?
When archival occurs it will first run on the active side and some time later it will run on the standby side as well. 
Before uploading history a check is done to see if it has already been uploaded, if so it is not re-uploaded.

### Can I specify multiple archival URIs?
No, each domain can only have one URI for history archival and one URI for visibility archival. Different domains, however, can have different URIs (with different schemes).

### How does archival work with PII?
No cadence workflow should ever operate on clear text PII. Cadence can be thought
of as a database and just as one would not store PII in a database PII should not be
stored in Cadence. This is even more important when archival is enabled because
these histories can be kept forever. 

## Planned Future Work
* Support archival of visibility.
* Support accessing histories without providing runID.
* Provide hard guarantee that no history is deleted from persistence before being archived if archival is enabled.
* Implement paused state. In this state no archivals will occur but histories also will not be deleted from persistence.
Once enabled again from paused state, all skipped archivals will occur. 
