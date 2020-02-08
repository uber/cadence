# Google Storage blobstore
## Configuration
See https://cloud.google.com/docs/authentication/production to understand how is made the authentication against google cloud storage

Nowdays we support three different ways in order to let Cadence know where your google keyfile credentials are located

* Cadence archival deployment.yaml configuration file
* `GOOGLE_APPLICATION_CREDENTIALS` environment variable
*  Google default credentials location

If more than one credentials location is given, then Cadence will resolve the conflicts by the following priority:

`GOOGLE_APPLICATION_CREDENTIALS > Cadencen archival deployment.yaml > Google default credentials`

Be sure that you have created your bucket first, and have enought rights in order to read/write over your bucket.

### Gcloud Archival example

Enabling archival is done by using the configuration below. `credentialsPath` is required but could be empty "" in order to allow a Google default credentials.

```
archival:
  history:
    status: "enabled"
    enableRead: true
    provider:
      gstorage:
        credentialsPath: "/tmp/keyfile.json"
  visibility:
    status: "enabled"
    enableRead: true
    provider:
      gstorage:
        credentialsPath: "/tmp/keyfile.json"

domainDefaults:
  archival:
    history:
      status: "enabled"
      URI: "gs://my-bucket-cad/cadence_archival/development"
    visibility:
      status: "enabled"
      URI: "gs://my-bucket-cad/cadence_archival/visibility"
```

## Visibility query syntax
You can query the visibility store by using the `cadence workflow listarchived` command

The syntax for the query is based on SQL

Supported column names are
- WorkflowID *String*
- StartTime *Date*
- CloseTime *Date*
- SearchPrecision *String - Day, Hour, Minute, Second*

WorkflowID and SearchPrecision are always required. One of StartTime and CloseTime are required and they are mutually exclusive.

Searching for a record will be done in times in the UTC timezone

SearchPrecision specifies what range you want to search for records. If you use `SearchPrecision = 'Day'`
it will search all records starting from `2020-01-21T00:00:00Z` to `2020-01-21T59:59:59Z` 

### Limitations

- The only operator supported is `=` 

### Example

*Searches for all records done in day 2020-01-21 with the specified workflow id*

`./cadence --do samples-domain workflow listarchived -q "StartTime = '2020-01-21T00:00:00Z' AND WorkflowID='workflow-id' AND SearchPrecision='Day'"`

## Archival query syntax

Once you have a workflowId and a runId you can retrieve your workflow history.

example:

`./cadence --do samples-domain --ct 15000  workflow  show  -w workflow-id -r runId`