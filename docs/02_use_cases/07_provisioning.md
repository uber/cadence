# Infrastructure Provisioning

Provisioning a new datacenter or a pool of machines in a public cloud is a potentially long running operation with
a lot of possibilities for intermittent failures. The scale is also a concern when tens or even hundreds of thousands of
 resources should be provisioned and configured. One useful feature for some of the provisioning scenarios is Cadence support for routing activity execution to a specific process or host.

A lot of operations require some sort of locking to ensure that no more than one mutation is executed on a resource at a time.
Cadence provides strong guarantees of uniqueness by business ID. This can be used to implement such locking behaviour in fault tolerant and scalable way.

Real life use case:

 * [Deployment to Kubernetes by Banzai Cloud](https://banzaicloud.com/blog/introduction-to-cadence/)
