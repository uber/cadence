# Archival

Archival is a feature that automatically moves histories from persistence to a blobstore after
the workflow retention period. The purpose of archival is to be able to keep histories as long as needed
while not overwhelming the persistence store. There are two reasons you may want
to keep the histories after the retention period has past:
1. **Compliance:** For legal reasons histories may need to be stored for a long period of time.
2. **Debugging:** Old histories can still be accessed for debugging.

Archival is still in beta and there are three limits to its feature set:
1. **Only Histories:** Only histories are archived, visibility records are simply deleted after the retention period.
2. **RunID Required:** In order to access an archived history, both workflowID and runID are required.
3. **Best Effort:** There are cases in which a history can be deleted from persistence without
being uploaded to blobstore first. These cases are rare but are possible with the current state of archival.

Work is being prioritized on archival to eliminate these limitations.

## Configuring Archival

Archival is controlled by both domain level config and cluster level config. 

### Cluster Archival Config

A Cadence cluster can be in one of three archival states:
  * **Disabled:** No archivals will occur and the blobstore is not initialized on startup.
  * **Paused:** This state is not yet implemented. Currently setting cluster to paused is the same as setting it to disabled.
  * **Enabled:** Archivals will occur.

Enabling the cluster for archival simply means histories are being written to blobstore. There is another
config which controls whether histories can be accessed from archival. Both these configs have defaults defined in static yaml, and have dynamic config overwrites.

### Domain Archival Config

A domain includes two pieces of archival related config: 
  * **Status:** Either enabled or disabled. If a domain is in the disabled state no archivals will occur for that domain. 
  A domain can safely switch between statuses.
  * **Bucket:** The location in blobstore where histories will be archived to. When a domain enables archival
  for the first time bucket is set and can never be mutated. If bucket is not specified when first enabling
  a domain for archival, a default bucket name from static config will be used.

## Running Locally

In order to run locally do the following:
1. `./cadence-server start`
2. `./cadence --do samples-domain domain register --gd false --archival_status enabled --retention 0`
3. Run the [helloworld cadence-sample](https://github.com/uber-common/cadence-samples) by following the README
4. Copy the workflowID and runID of the completed workflow from log output
5. `./cadence --do samples-domain wf show --wid <workflowID> --rid <runID>`

This archival setup is using the local filesystem to implement the blobstore. You can see the blobstore is mounted
out of the `/tmp/development/blobstore` directory. 

Cadence also supports running locally with S3. Follow the README in `common/blobstore/s3store/README.md`.

## FAQ

### How does archival interact with global domains?
When archival occurs it will first run on the active side and some time later it will run on the standby side as well. 
Before uploading history a check is done to see if it has already been uploaded, if so it is not re-uploaded.

### What are the supported blobstores?
All archival code interacts with a blobstore interface. Currently there 
are two implementations of this interface: S3 and filestore. Filestore is ony used
for running locally. In order to run on top of a different blobstore all that needs
to be done is to write a new implementation of the interface backed by the desired blobstore.

### Why are .index files uploaded along with history files?
Global domains can have multiple versions of histories due to conflict resolution. Each
archival task operates over a single version of history. The .index file contains all versions that have been uploaded
for a single execution.

History blob names include version as part of the name. This means that archival tasks
of different versions will not upload conflicting blobs. 

### Why not just use list on history blobs to get all the versions?
Since blob names contain version it is reasonable to use ```ListByPrefix``` to get all 
versions of archived history. However some implementations of blobstore do not support ```ListByPrefix```.
For these implementations, uploading an index file allows clients to determine which versions are available.
There is an open work item which will separate the implementations for listable vs non-listable blobstores.

### Why are multiple blobs uploaded for a single history?
Some blobstores do not support a seek API. For these blobstores the only way to do pagination over history is to upload multiple
blobs with increasing page numbers.

## Planned Future Work
* Support accessing histories without providing runID.
* Support archival of visibility.
* Provide hard guarantee that no history is deleted from persistence before being archived if archival is enabled.
* Implement paused state. In this state no archivals will occur but histories also will not be deleted from persistence.
Once enabled again from paused state, all skipped archivals will occur. 
* Separate implementations based on listable blobstores from those based on non-listable blobstores.