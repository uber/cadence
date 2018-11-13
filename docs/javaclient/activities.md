---
layout: doc
title: Activities
weight: 31
---

# Activities

An activity is a manifestation of a particular task in the business logic.

## Activity Interface

Activities are defined as methods of a plain Java interface. Each method defines a single activity type. A single
workflow can use more than one activity interface and call more that one activity method from the same interface.
The only requirement is that activity method arguments and return values are serializable to a byte array using the provided
[DataConverter](src/main/java/com/uber/cadence/converter/DataConverter.java) interface. The default implementation uses
JSON serializer, but an alternative implementation can be easily configured.

Example of an interface that defines four activities:

```java
public interface FileProcessingActivities {

    void upload(String bucketName, String localName, String targetName);

    String download(String bucketName, String remoteName);

    @ActivityMethod(scheduleToCloseTimeoutSeconds = 2)
    String processFile(String localName);

    void deleteLocalFile(String fileName);
}

```
An optional @ActivityMethod annotation can be used to specify activity options like timeouts or a task list. Required options
that are not specified through the annotation must be specified at run time.

## Activity Implementation

Activity implementation is an implementation of an activity interface. A single instance of the activities implementation
is shared across multiple simultaneous activity invocations. Therefore, the activity implementation code must be *thread safe*.

The values passed to activities through invocation parameters or returned through a result value are recorded in the execution history.
The entire execution history is transferred from the Cadence service to workflow workers when a workflow state needs to recover.
A large execution history can thus adversely impact the performance of your workflow.
Therefore, be mindful of the amount of data you transfer via activity invocation parameters or return values.
Other than that, no additional limitations exist on activity implementations.

```java
public class FileProcessingActivitiesImpl implements FileProcessingActivities {

    private final AmazonS3 s3Client;

    private final String localDirectory;

    void upload(String bucketName, String localName, String targetName) {
        File f = new File(localName);
        s3Client.putObject(bucket, remoteName, f);
    }

    String download(String bucketName, String remoteName, String localName) {
        // Implementation omitted for brevity.
        return downloadFileFromS3(bucketName, remoteName, localDirectory + localName);
    }

    String processFile(String localName) {
        // Implementation omitted for brevity.
        return compressFile(localName);
    }

    void deleteLocalFile(String fileName) {
        File f = new File(localDirectory + fileName);
        f.delete();
    }
}
```
### Accessing Activity Info

The [Activity](src/main/java/com/uber/cadence/activity/Activity.java) class provides static getters to access information about the workflow that invoked it.
Note that this information is stored in a thread local variable. Therefore, calls to Activity accessors succeed only in the thread that invoked the activity function.

```java
public class FileProcessingActivitiesImpl implements FileProcessingActivities {

     @Override
     public String download(String bucketName, String remoteName, String localName) {
        log.info("domain=" +  Activity.getDomain());
        WorkflowExecution execution = Activity.getWorkflowExecution();
        log.info("workflowId=" + execution.getWorkflowId());
        log.info("runId=" + execution.getRunId());
        ActivityTask activityTask = Activity.getTask();
        log.info("activityId=" + activityTask.getActivityId());
        log.info("activityTimeout=" + activityTask.getStartToCloseTimeoutSeconds());
        return downloadFileFromS3(bucketName, remoteName, localDirectory + localName);
     }
     ...
 }
```

### Asynchronous Activity Completion

Sometimes an activity lifecycle goes beyond a synchronous method invocation. For example, a request can be put in a queue
and later a reply comes and is picked up by a different worker process. The whole request-reply interaction can be modeled
as a single Cadence activity.

To indicate that an activity should not be completed upon its method return, call Activity.doNotCompleteOnReturn() from the
original activity thread.
Then later, when replies come, complete the activity using [ActivityCompletionClient](src/main/java/com/uber/cadence/client/ActivityCompletionClient.java).
To correlate activity invocation with completion use either `TaskToken` or workflow and activity IDs.

```java
public class FileProcessingActivitiesImpl implements FileProcessingActivities {

     public String download(String bucketName, String remoteName, String localName) {
         byte[] taskToken = Activity.getTaskToken(); // Used to correlate reply
         asyncDownloadFileFromS3(taskToken, bucketName, remoteName, localDirectory + localName);
         Activity.doNotCompleteOnReturn();
         return "ignored"; // Return value is ignored when doNotCompleteOnReturn was called.
     }
     ...
}
```
When download is complete, the download service potentially calls back from a different process:
```java
    public <R> void completeActivity(byte[] taskToken, R result) {
        completionClient.complete(taskToken, result);
    }

    public void failActivity(byte[] taskToken, Exception failure) {
        completionClient.completeExceptionally(taskToken, failure);
    }
```

### Activity Heartbeating

Some activities are long running. To react to their crashes quickly, use a heartbeat mechanism.
Use the `Activity.heartbeat` function to let the Cadence service know that the activity is still alive. You can piggyback
`details` on an activity heartbeat. If an activity times out, the last value of `details` is included
in the ActivityTimeoutException delivered to a workflow. Then the workflow can pass the details to
the next activity invocation. This acts as a periodic checkpointing mechanism of an activity progress.
```java
public class FileProcessingActivitiesImpl implements FileProcessingActivities {

     @Override
     public String download(String bucketName, String remoteName, String localName) {
        InputStream inputStream = openInputStream(file);
        try {
            byte[] bytes = new byte[MAX_BUFFER_SIZE];
            while ((read = inputStream.read(bytes)) != -1) {
                totalRead += read;
                f.write(bytes, 0, read);
                /*
                 * Let the service know about the download progress.
                 */
                 Activity.heartbeat(totalRead);
            }
        } finally {
            inputStream.close();
        }
     }
     ...
}
```
