# Overview

Cadence is a task orchestrator for your application's tasks. Applications using Cadence can execute
a logical flow of tasks (especially long-running business logic) asynchronously or synchronously, 
and can scale runtime on distributed systems without you, the service owner, worrying about 
infrastructure needs.

A quick example illustrates its use case. Consider Uber Eats where Cadence manages the entire 
business flow from placing an order, accepting it, handling shopping cart processes (adding, 
updating, and calculating cart items), entering the order in a pipeline (for preparing food and 
coordinating delivery), to scheduling delivery as well as handling payments.

Cadence consists of a programming framework (or client library) and a managed service (or backend).
The framework enables developers to author and coordinate tasks in familiar languages. Initially, 
as the framework only supports Go, code samples in this guide focus on Go.

[Watch Maxim's talk](https://youtu.be/llmsBGKOuWI) from the Uber Open Summit for an introduction 
to the Cadence programming model and value proposition.

The GitHub repo for the Cadence server is [uber/cadence](https://github.com/uber/cadence). The docker 
image for the Cadence server is available on Docker Hub at 
[ubercadence/server](https://hub.docker.com/r/ubercadence/server).
