# Overview

A large number of use cases span beyond a single request-reply, require tracking
of a complex state, respond to asynchronous events, and communicate to external unreliable dependencies.
The usual approach to building such applications is a hodgepodge of stateless services,
databases, cron jobs and queuing systems. This negatively impacts the developer productivity as most of the code is
dedicated to plumbing, obscuring the actual business logic behind myriad of low level details. Such systems frequently have availability problems as it is hard to keep all the components healthy.

The Cadence solution is a _fault-oblivious stateful_ programming model that hides most of the complexities of building scalable distributed applications. In essence, Cadence provides a durable virtual memory that is not
linked to a specific process and preserves the full application state including function stacks with local variables across all sort of host and software failures.
This allows to write code using the full power of a programming language while Cadence takes care of durability, availability and scalability of the application.

Cadence consists of a programming framework (or client library) and a backend service.

The framework enables developers to author fault-oblivious code in familiar languages
([Go](https://github.com/uber-go/cadence-client/) and [Java](https://github.com/uber/cadence-java-client)
are supported today with some projects in [Python](https://github.com/firdaus/cadence-python) and
[C#](https://github.com/nforgeio/neonKUBE/tree/master/Lib/Neon.Cadence)
via [proxy](https://github.com/nforgeio/neonKUBE/tree/master/Go/src/github.com/loopieio/cadence-proxy)
in development).

The backend service is stateless and relies on a persistent store. Currently Cassandra and MySQL stores
are supported. Adapter to any other database that provides multirow single shard transactions
can be added. There are different service deployment models. At Uber, our team operates multitenant clusters
that are shared by hundreds of applications.

[Watch Maxim's talk](https://youtu.be/llmsBGKOuWI) from the Uber Open Summit for an introduction
to the Cadence programming model and value proposition.

The GitHub repo for the Cadence server is [uber/cadence](https://github.com/uber/cadence). The docker
image for the Cadence server is available on Docker Hub at
[ubercadence/server](https://hub.docker.com/r/ubercadence/server).
