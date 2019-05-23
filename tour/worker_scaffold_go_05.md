---
codecontent: workerscaffoldgo05
weight: 40
categories: [tour]
---

# Testing the Sample Helper

In the root directory of the project, use a text editor to create a **main.go** file with the
following code. This code tests that the Worfklow Service Client can be successfully created. 

To execute the worker, run the following commands from the root directory:

```bash
dep ensure
go build
./cadence-scaffold
```

Note that you don't need to have the local Cadence service running to execute this test.
