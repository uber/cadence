# Periodic Execution (aka Distributed Cron)

Execute some business logic periodically. The advantage of Cadence in this scenario is that it provides 
guarantees of the execution, sophisticated error handling and retry policies. 

Another important dimension is scale. Some use cases require periodic execution for large number of entities. 
At Uber there are applications that create periodic workflow per customer. 
Think about 100+ million parallel cron jobs that don't require any separate batch processing framework.   
