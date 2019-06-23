# Periodic Execution (aka Distributed Cron)

Execute some business logic periodically. The advantage of Cadence in this scenario is that it provides
guarantees of the execution, sophisticated error handling, retry policies and visibility into execution history.

Another important dimension is scale. Some use cases require periodic execution for large number of entities.
At Uber there are applications that create periodic workflow per customer.
Think about 100+ million parallel cron jobs that don't require any separate batch processing framework.

Periodic execution is frequently part of other use cases. For example once a month report generation is a periodic service orchestration. Or a event driven workflow that accumulates loyalty points for a customer may apply those points once a month.

The couple (from dozens) real life examples of Cadence periodic executions:

 * Hexaggregator is a Uber backend service that recalculates various statistics for each [hex](https://eng.uber.com/h3/) in each city once a minute.
 * Monthly Uber for Business report generation