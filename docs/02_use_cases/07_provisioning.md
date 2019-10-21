# Infrastructure Provisioning

Provisioning a new datacenter or a pool of machines in a public cloud is a potentially long running operation with
a lot of possibilities for intermittent failures. The scale is also a concern when tens or even hundreds of thousands of resources should be provisioned and configured. One useful feature for provisioning scenarios is Cadence support for routing activity execution to a specific process or host.

A lot of operations require some sort of locking to ensure that no more than one mutation is executed on a resource at a time.
Cadence provides strong guarantees of uniqueness by business ID. This can be used to implement such locking behavior in a fault tolerant and scalable manner.

Some real-world use cases:

 * [Using Cadence workflows to spin up Kubernetes, by Banzai Cloud](https://banzaicloud.com/blog/introduction-to-cadence/)
 * [Using Cadence to orchestrate cluster life cycle in HashiCorp Consul, by HashiCorp](https://www.youtube.com/watch?v=kDlrM6sgk2k&feature=youtu.be&t=1188)
