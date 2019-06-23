# Microservice Orchestration and Saga

It is common that some business process is implemented as multiple microservice calls.
And the implementation must guarantee that all of the calls must eventually succeed even in presence of prolonged downstream service failures.
In some cases instead of trying to complete the process by retrying for a long time some compensation rollback logic should be executed.
[Saga Pattern](https://microservices.io/patterns/data/saga.html) is one way to standardise on compensation APIs.

Cadence is a perfect fit for such scenarios. It guarantees that workflow code eventually completes, has built it support
for unlimited exponential activity retries and simplifies coding of the compensation logic. It also gives full visibility into the state of each workflow. Contrast it with an orchestration based on queues where getting a current status of each invividual request is practically imposisble.

The real life examples of Cadence based service orchestration scenarios:

 * [Deployment to Kubernetes by Banzai Cloud](https://banzaicloud.com/blog/introduction-to-cadence/)
 * [Uber Customer Obsession Platform](https://eng.uber.com/customer-obsession-ticket-routing-workflow-and-orchestration-engine/)



