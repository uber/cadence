# Durable Functions aka Workflows

## Overview

Cadence core abstraction is a **durable workfow function**. The state of the function including local variables and threads it creates is immune to process failures.
This is very powerful concept as it encapsulates state, processing threads, durable timers and event handlers.

## Example

Let's look at a use case. A customer signs up to an application with a trial period. After the period if the customer is not cancelled he should be charged once a month for the renewal. Customer has to be notified by email about the charges and should be able to cancel the subscription at any time.

The business logic of the use case is not very complicated and can be expressed in a few dozen lines of code. But any practical implementation has to ensure that the business process is fault tolerant and scalable. There are various ways to approach the design of such system.

One apporach is to center it around a database. An application processes would periodically scan database tables for customers in specific states, execute necessary actions and update the state to reflect that. While feasible this apporach has various drawbacks. Most obvious one is that the state machine of the customer state quickly becomes extremely complicated. For example a credit card charge or emails sending can fail due to downstream system unavailability. The failed calls should be retried, ideally using an exponential retry policy. These calls should be throttled to not overload external systems. There should be support for poison pills to avoid blocking the whole process if a single customer record cannot be processed for whatever reason. Database based apporach also usually has peformance problems. Databases are not efficient for scenarios that require constant polling for records in a specific state.

Another commonly employed approach is to use a timer service and queues. So any update is pushed to a queue and then a worker that consumes from it updates a database and possibly pushes more messages in downstream queues. For operations that require scheduling an external timer service can be used. This apporach usually scales much better as a database is not constantly polled for changes. But it makes the programming model even more complex and error prone as usually there is no transactional update between a queing system and a database.

With Cadence the whole logic can be encapsulated in a simple durable function that directly implements the business logic. As the function is stateful the implementer doesn't need to employ any additional systems to ensure durability and fault tolerance.

Here is an example workflow that implements the subscription management use case. It is in Java, but Go is also supported. The Python and .NET libraries are under active development.

```java
public interface SubscriptionWorkflow {
    @WorkflowMethod
    void execute(String customerId);
}

public class SubscriptionWorkflowImpl implements SubscriptionWorkflow {

  private final SubscriptionActivities activities =
      Workflow.newActivityStub(SubscriptionActivities.class);

  @Override
  public void execute(String customerId) {
    activities.sendWelcomeEmail(customerId);
    try {
      boolean trialPeriod = true;
      while (true) {
        Workflow.sleep(Duration.ofDays(30));
        activities.chargeMonthlyFee(customerId);
        if (trialPeriod) {
          activities.sendEndOfTrialEmail(customerId);
          trialPeriod = false;
        } else {
          activities.sendMonthlyChargeEmail(customerId);
        }
      }
    } catch (CancellationException e) {
      activities.processSubscriptionCancellation(customerId);
      activities.sendSorryToSeeYouGoEmail(customerId);
    }
  }
}
```
Again, note that this code directly implements the business logic. If any of the invoked operations (aka activities) takes long time the code is not going to change. It is OK to block on `chargeMonthlyFee` for a day if the downstream processing service is down that long. The same way the blocking sleep for 30 days is a normal operation inside the workflow code.

Cadence has practically no scalability limits on number of open workflow instances. So even if your site has hundreds of millions of consumers the above code is not going to change.

The commonly asked question by the developers that learn Cadence is "How do I handle workflow worker process failure/restart in my workflow"? The answer is that you do not. The workflow code is completely oblivious to any failures and downtime of workers or even Cadence service itself. As soon as they are recovered and the workflow needs to handle some event like timer or an activity completion the current state of the workflow is fully restored and continues execution. The only reason for the workflow failure is workflow business code throwing an exception, not underlying infrasturcture outages.

## State Recovery and Determinism

The workflow function state recovery utilizes event sourcing which puts a few restrictions on how the code is written. The main restriction is that the workflow code must be deterministic which means that it must produce exactly the same result if executed multiple times. It rules out any external API calls from the workflow code as external calls can fail intermittently or change its output any time. That is why all communication with external world should happen through activities. For the same reason workflow code must use Cadence APIs to get current time, sleep and create new threads.

To understand the Cadence execution model as well as the recovery mechanism watch [this webcast](https://youtu.be/qce_AqCkFys). The animation covering recovery starts at 15:50.

## ID Uniqueness

Workflow ID is assigned by a client when starting a workflow. It is usually a business level ID like customer ID or order ID.

Cadence guarantees that there could be only one workflow (across all workflow types) with given ID open per domain at any time. An attempt to start workflow with the same ID is going to fail with WorkflowExecutionAlreadyStarted error.

An attempt to start a workflow if there is a completed workflow with the same ID depends on a `WorkflowIdReusePolicy` option:

- `AllowDuplicateFailedOnly` means that it is allowed to start a workflow only if a previously run workflow with the same ID failed.
- `AllowDuplicate` means that it is allowed to start independently of the previous workflow completion status.
- `RejectDuplicate` mans that it is not allowed start a workflow execution using the same workflow ID at all.

The default is `AllowDuplicateFailedOnly`.

To distinguish multiple runs of a workflow with the same workflow ID Cadence identifies a workflow by two ids: `Workflow ID` and `Run ID`. `Run ID` is a service assigned UUID. To be precise any workflow is uniquely identified by a triple: `Domain Name`, `Workflow ID` and `Run ID`.

## Child Workflow

A workflow can execute other workflows as `child workflows`. A child workflow completion or failure is reported to its parent.

Here are reasons to use child workflows:

- A child workflow can be hosted by a separate set of workers which don't contain the parent workflow code. So it would act as a separate service that can be invoked from multiple other workflows.
- A single workflow has a limited size. For example it cannot execute 100k activities. Child workflows can be used to partition the problem into smaller chunks. One parent with 1000 children each executing 1000 activities gives 1 million activities executed.
- A child workflow can be used to manage some resource using its ID to guarantee uniqueness. For example a workflow that manages host upgrades can have a child workflow per host (host name being a workflow ID) and use them to ensure that all operations on the host are serialized.
- A child workflow can be used to execute some periodic logic without blowing up the parent history size. Parent starts a child which executes periodic logic calling continue as new as many times as needed, then completes. From the parent point if view it is just a single child workflow invocation.

The main limitation of a child workflow versus collocating all the application logic in a single workflow is lack of the shared state. Parent and child can communicate only through asynchronous signals. But if there is a tight coupling between them it might be simpler to use a single workflow and just rely on a shared object state.

It is recommended starting from a single workflow implementation if your problem has bounded size in terms of number of executed activities and processed signals. It is simpler than multiple asynchronously communicating workflows.

## Workflow Retries

Workflow code is unaffected by infrastructure level downtime and failures. But it still can fail due to business logic level failures. For example an activity can fail due to exceeding retry interval and error is not handled by appliation code or workflow code has a bug.

Some workflows require guarantee that they keep running even in presence of such failures. To support such use cases an optional exponential retry policy can be specified when starting a workflow. When specified a workflow failure restarts a workflow from the beginning after the calculated retry interval. Here are retry policy parameters:

- InitialInterval is a delay before the first retry
- BackoffCoefficient. Retry policies are exponential. The coefficient specifies how fast the retry interval is growing. The coefficient of 1 means that retry interval is always equal to the InitialInterval.
- MaximumInterval specifies the maximum interval between retries. Useful for coefficients more than 1.
- MaximumAttempts specifies how many times to attempt to execute a workflow in presence of failures. If this limit is exceeded the workflow fails without retry. Not required if ExpirationInterval is specified.
- ExpirationInterval specifies for how long to attempt executing an workflow in presence of failures. If this interval is exceeded the workflow fails without retry. Not required if MaximumAttempts is specified.
- NonRetryableErrorReasons allows to specify erros that shouldn't be retried. For example retrying invalid arguments error doesn't make sense in some scenarios.

