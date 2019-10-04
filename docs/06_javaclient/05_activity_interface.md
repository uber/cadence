# Activity Interface

An activity is a manifestation of a particular task in the business logic.

Activities are defined as methods of a plain Java interface. Each method defines a single activity type. A single
workflow can use more than one activity interface and call more that one activity method from the same interface.
The only requirement is that activity method arguments and return values are serializable to a byte array using the provided
[DataConverter](https://static.javadoc.io/com.uber.cadence/cadence-client/2.4.1/index.html?com/uber/cadence/converter/DataConverter.html)
interface. The default implementation uses a JSON serializer, but an alternative implementation can be easily configured.

Following is an example of an interface that defines four activities:

```java
public interface FileProcessingActivities {

    void upload(String bucketName, String localName, String targetName);

    String download(String bucketName, String remoteName);

    @ActivityMethod(scheduleToCloseTimeoutSeconds = 2)
    String processFile(String localName);

    void deleteLocalFile(String fileName);
}

```
We recommend to use a single value type argument for activity methods. In this way, adding new arguments as fields
to the value type is a backwards-compatible change.

An optional @ActivityMethod annotation can be used to specify activity options like timeouts or a task list. Required options
that are not specified through the annotation must be specified at runtime.
