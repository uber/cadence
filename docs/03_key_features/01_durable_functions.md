# Durable Functions aka Workflows

Cadence core abstraction is a durable workfow function. The state of the function including local variables and threads it creates is immune to process failures.
This is very powerful concept as it encapsulates state, processing threads, durable timers and event handlers.

Let's look at a use case. A customer signs up to an application with a trial period. After the period if the customer is not cancelled he should be charged once a month for the renewal. Customer has to be notified by email about the charges and should be able to cancel the subscription at any time.

The business logic of the use case is not very complicated and can be expressed in a few dozen lines of code. But any practical implementation has to ensure that the business process is fault tolerant and scalable. There are various ways to approach the design of such system.

One apporach is to center it around a database. An application processes would periodically scan database tables for customers in specific states, execute necessary actions and update the state to reflect that. While feasible this apporach has various drawbacks. Most obvious one is that the state machine of the customer state quickly becomes extremely complicated. For example credit card charge or emails sending can fail due to downstream system unavailability. The failed calls should be retried, ideally using an exponential retry policy. These calls should be throttled to not overload the external systems. There should be support for poison pills to avoid blocking the whole process if a single customer record cannot be processed for whatever reason. Database based apporach also usually has peformance problems. Databases are not efficient for schenarios that require constant polling for records in a specific state.

Another commonly used approach is to use a timer service and a queue. So any update is pushed to a queue and then a worker that consumes from it updates a database and possibly pushes more messages in downstream queues. For operations that require scheduling an external timer service can be used. This apporach usually scales much better as a database is not constantly polled for changes. But it makes the programming model even more complex and error prone as usually there is no transactional update between a queing system and a database.

With Cadence the whole logic can be encapsulated in a simple durable function that directly implements the business logic. As the function is stateful the implementer doesn't need to employ any additional systems to ensure durability and fault tolerance.

Here is an example workflow that implements the subscription management use case:

{% codetabs %}

{% codetab Java %}
```java
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
{% endcodetab %}
{% codetab Go %}
```go
	EncodedValue struct {
		value         []byte
		dataConverter encoded.DataConverter
	}
	// Version represents a change version. See GetVersion call.
	Version int
```
{% endcodetab %}
{% endcodetabs %}
Foop
