---
codecontent: workerscaffoldgo03
weight: 30
categories: [tour]
---

# Create a Helper for Building Samples

Cadence workflows and activities run in workers. The workers poll the Cadence service for tasks, execute the 
workflow or activity code, and respond with decisions back to the Cadence service. The complexity of this 
interaction is handled by the Cadence client that we added as a dependency in the previous step. 

In this step, we'll create a helper for samples in this tour. This helper will start as a barebones framework 
for creating a worker. 

First, create a subdirectory called **common** in your Go project. Using a text editor, create 
**sample_helper.go** and add the following code.

This sets up Zap for development logging and Tally to not report metrics. The code is not ready to build yet. 
In the next step, you'll build the Cadence workflow service client.
