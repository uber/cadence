# Use Cases
Cadence already supports dozens of production applications at Uber and at external companies.
This page tries to categorize the types of use cases that are good fit for the platform. 
[Contact the development team](https://join.slack.com/t/uber-cadence/shared_invite/enQtNDczNTgxMjYxNDEzLTI5Yzc5ODYwMjg1ZmI3NmRmMTU1MjQ0YzQyZDc5NzMwMmM0NjkzNDE5MmM0NzU5YTlhMmI4NzIzMDhiNzFjMDM)
if you have question about your specific requirements.

## Periodic Execution (aka Distributed Cron)

Execute some business logic periodically. The advantage of Cadence in this scenario is that it provides 
guarantees of the execution, sophisticated error handling and retry policies. 

Another important dimension is scale. Some use cases require periodic execution for large number of entities. 
At Uber there are applications that create periodic workflow per customer. 
Think about 100+ million parallel cron jobs that don't require any separate batch processing framework.   

## Microservice Orchestration and Saga

It is common that some business process is implemented as multiple microservice calls. 
And the implementation must guarantee that all of the calls must eventually succeed even in presence of prolonged downstream service failures.
In some cases instead of trying to complete the process by retrying for a long time some compensation rollback logic should be executed.
[Saga Pattern](https://microservices.io/patterns/data/saga.html) is one way to standardise on compensation APIs.

Cadence is a perfect fit for such scenarios. It guarantees that workflow code eventually completes, has built it support
for unlimited exponential activity retries and simplifies coding of the compensation logic. 

## Event Driven Application

Many applications listen to multiple event sources, update state of correspondent business entities 
and have to execute some actions if some state is reached.
Cadence is a good fit for many of them. It has direct support for asynchronous events (aka signals), 
has a simple programming model that hides a lot of complexity
around state persistence and ensures external action execution through built-in retries. 

## Storage Scan

It is common to have large data sets partitioned across large number of hosts or databases or having billions of files in an S3 bucket.
Cadence is an ideal solution for implementing the full scan of such data in a scalable and resilient way. The standard pattern 
is to run an activity (or multiple parallel activities for partitioned data sets) that performs the scan and heartbeats its progress
back to Cadence. In a case of a host failure the activity is retried on a different host and continues execution from the last reported progress. 

## Batch Job

A lot of batch jobs are not pure data manipulation programs. For those the existing big data frameworks are the best fit.
But if processing a record requires external API calls that might fail and potentially take long time Cadence might be a better fit.
One of internal Uber customers uses Cadence for end of month statement generation. Each statement requires calls to multiple 
 microservices and some statements can be really large. Cadence was choosen as it provides hard guarantees around durability of the financial data
 and seamlessly deals with long running operations, retries and other intermittent failures. 

## Infrastructure Provisioning

Provisioning a new datacenter or a pool of machines in a public cloud is a potentially long running operation with 
a lot of possibilities for intermittent failures. The scale is also a concern when tens or even hundreds of thousands of
 resources should be provisioned and configured.
A lot of operations require some sort of locking to ensure that no more than one mutation is executed on a resource at a time. 
Cadence provides strong guarantees of uniqueness by business ID. This can be used to implement such locking behaviour in fault tolerant and scalable way. 

## Deployment

Deployment of applications to containers or virtual or physical machines is a non trivial process. 
Its business logic has to deal with complex requirements around rolling upgrades, canary deployments and rollbacks.
Cadence is a perfect platform for building a deployment solution as it provides all the necessary guarantees and abstractions 
allowing developers focusing on the business logic.

## Operational Management

Imagine that you have to create a self operating database similar to Amazon RDS. Cadence is used in multiple projects 
that automate managing and automatic recovery of various products like MySQL, Elastic Search and Cassandra. 

## Interactive Application

Cadence is performant and scalable enough to support interactive applications. It can be used to track UI session state and
at the same time execute background operations. For example while placing an order a customer might need to go through several screens
while a background task evaluates the customer for fraudulent activity.

## DSL Workflows

Cadence supports implementing business logic directly in a programming language like Java and Go. But there are cases when 
using a domain specific language is more appropriate. Or there is a legacy system that uses some form of DSL for process definition
but it is not operationally stable and scalable. 
This also applies to not so legacy systems like Airflow, various BPMN engines and AWS Step Functions. 

An application that interprets the DSL definition can be written using Cadence SDK. It automatically becomes highly fault tolerant,
scalable and durable when runs on Cadence. 

Cadence was already used to deprecate several Uber internal  DSL engines. The customers continue to use existing
process definitions, but Cadence is used as an execution engine. 

There are multiple benefits of unifying all company workflow engines on top of Cadence. The most obvious one is that
it is more efficient to support a single product instead of many. It is also hard to beat scalability and stability of 
Cadence which each of the integrations gets practically for free. An ability to share activities across "engines" 
might be a big benefit in some cases.

## Big Data and ML

A lot of companies build custom ETL and ML training and deployment solutions. Cadence is a good fit for a control plain for such applications.
One important feature of Cadence is ability to route task execution to specific process or host. 
It is useful to control how ML models and other large files are allocated to hosts. 
For example if ML model is partitioned by city the requests should be routed to hosts that contain the corresponding city model.