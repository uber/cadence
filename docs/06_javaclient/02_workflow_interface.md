# Workflow Interface

Workflow encapsulates the orchestration of activities and child workflows.
It can also answer synchronous queries and receive external events (also known as signals).

A workflow must define an interface class. All of its methods must have one of the following annotations:

- **@WorkflowMethod** indicates an entry point to a workflow. It contains parameters such as timeouts and a task list.
  Required parameters (such as `executionStartToCloseTimeoutSeconds`) that are not specified through the annotation must be provided at runtime.
- **@SignalMethod** indicates a method that reacts to external signals. It must have a `void` return type.
- **@QueryMethod** indicates a method that reacts to synchronous query requests.

You can have more than one method with the same annotation. For example:
```java
public interface FileProcessingWorkflow {

    @WorkflowMethod(executionStartToCloseTimeoutSeconds = 10, taskList = "file-processing")
    String processFile(Arguments args);

    @QueryMethod(name="history")
    List<String> getHistory();

    @QueryMethod(name="status")
    String getStatus();

    @SignalMethod
    void retryNow();
}
```

We recommended that you use a single value type argument for workflow methods. This way, adding new arguments as fields to the value type is a backwards-compatible change.
