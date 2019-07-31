# Periodic Execution (aka Distributed Cron)

Periodic execution, frequently referred to as distributed cron, is when you execute business logic periodically. The advantage of Cadence for these scenarios is that it guarantees execution, sophisticated error handling, retry policies, and visibility into execution history.

Another important dimension is scale. Some use cases require periodic execution for a large number of entities.
At Uber, there are applications that create periodic workflows per customer.
Imagine 100+ million parallel cron jobs that don't require a separate batch processing framework.

Periodic execution is often part of other use cases. For example, once a month report generation is a periodic service orchestration. Or an event-driven workflow that accumulates loyalty points for a customer and applies those points once a month.

There are many real-world examples of Cadence periodic executions. Such as the following:

 * An Uber backend service that recalculates various statistics for each [hex](https://eng.uber.com/h3/) in each city once a minute.
 * Monthly Uber for Business report generation.
