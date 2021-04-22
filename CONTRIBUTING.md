# Developing Cadence

This doc is intended for contributors to `cadence` server (hopefully that's you!)

Join our Slack channel(invite link in the [home page](https://github.com/uber/cadence#cadence)) #development if you need help. 
>Note: All contributors need to fill out the [Uber Contributor License Agreement](http://t.uber.com/cla) before we can merge in any of your changes

## Development Environment

* Golang. Install on OS X with.
```
brew install go
``` 
* Make sure set PATH to include bin path of GO so that other executables like thriftrw can be found.
```bash
# check it first
echo $GOPATH
# append to PATH
PATH=$PATH:$GOPATH/bin
# to confirm, run
echo $PATH
``` 

* Download golang dependencies. 
```bash
go mod download
```

After check out and go to the Cadence repo, compile the `cadence` service and helper tools without running test:

```bash
make bins
``` 

>Note: If running into any compiling issue, 
>1. Make sure you upgrade to the latest stable version of Golang.
> 2. Check if this document is outdated by comparing with the building steps in [Dockerfile](https://github.com/uber/cadence/blob/master/Dockerfile)
* Database. The default setup of Cadence depends on Cassandra. Before running the Cadence or tests you must have `cassandra` dependency(or equivalent database in the below notes)

>Note: This section assumes you are working with Cassandra. Please refer to [persistence documentation](https://github.com/uber/cadence/blob/master/docs/persistence.md) if you want to test with others like MySQL/Postgres.
> Also, you don't need those local stores if you want to connect to your existing staging/QA environment. 

```bash
# install cassandra
# you can reduce memory used by cassandra (by default 4GB), by following instructions here: http://codefoundries.com/developer/cassandra/cassandra-installation-mac.html
brew install cassandra

# start services
brew services start cassandra 
# or 
cassandra -f

# after a short while, then use cqlsh to login, to make sure Cassandra is up  
cqlsh 
Connected to Test Cluster at 127.0.0.1:9042.
[cqlsh 5.0.1 | Cassandra 3.11.8 | CQL spec 3.4.4 | Native protocol v4]
Use HELP for help.
cqlsh>
```
>If you are running Cadence on top of [Mysql](docs/setup/MYSQL_SETUP.md) or [Postgres](docs/setup/POSTGRES_SETUP.md), you can follow the instructions to run the SQL DB and then run:

Now you can setup the database schema
```bash
make install-schema

# If you use SQL DB, then run:
make install-schema-mysql OR make install-schema-postgres
```

Then you will be able to run a basic local Cadence server for development:
```bash
./cadence-server start

# If you use SQL DB, then run:
./cadence-server --zone <mysql/postgres> start
```

You can run some workflow [samples](https://github.com/uber-common/cadence-samples) to test the development server.
 
> This basic local server doesn't have some features like advanced visibility, archival, which require more dependency than Cassandra/Database and setup. 

## Issues to start with

Take a look at the list of issues labeled with
[good first issue](https://github.com/uber/cadence/labels/good%20first%20issue).
These issues are a great way to start contributing to Cadence.

Later when you are more familiar with Cadence, look at issues with 
[up-for-grabs](https://github.com/uber/cadence/labels/up-for-grabs). 



## Repository layout
A Cadence server cluster is composed of four different services: Frontend, Matching, History and Worker(system).
Here's what's in each top-level directory in this repository:

* **canary/** : The test code that needs to run periodically to ensure Cadence is healthy
* **client/** : Client wrappers to let the four different services talk to each other
* **cmd/** : The main function to build binaries for servers and CLI tools
* **config/** : Sample configuration files
* **docker/** : Code/scripts to build docker images
* **docs/** : Documentation
* **host/** : End-to-end integration tests
* **schema/** : Versioned persistence schema for Cassandra/MySQL/Postgres/ElasticSearch
* **scripts/** : Scripts for CI build
* **tools/** : CLI tools for Cadence workflows and also schema updates for persistence
* **service/** : Contains four sub-folders that dedicated for each of the four services
* **common/** :  Basically contains all the rest of the code in Cadence server, the names of the sub folder are the topics of the packages

If looking at the Git history of the file/package, there should be some engineers focusing on each area/package/component that you are working on.
You will probably get code review from them. Don't hesitate to ask for some early feedback or help on Slack.

## Testing & Debug

To run all the package tests use the below command.
This will run all the tests excluding end-to-end integration test in host/ package):

```bash
make test
```
:warning: Note:
> You will see some test failures because of errors connecting to MySQL/Postgres if only Cassandra is up. This is okay if you don't write any code related to persistence layer. 
 
To run all end-to-end integration tests in **host/** package: 
```bash
make test_e2e
```

To debug a specific test case when you see some failure, you can trigger it from an IDE, or use the command
```
go test -v <path> -run <TestSuite> -testify.m <TestSpercificTaskName>
# example:
go test -v github.com/uber/cadence/common/persistence/persistence-tests -run TestVisibilitySamplingSuite -testify.m TestListClosedWorkflowExecutions
```

## Pull Requests
After all the preparation you are about to write code and make a Pull Request for the issue.

### When to open a PR
You have a few options for choosing when to submit:

* You can open a PR with an initial prototype with "Draft" option or with "WIP"(work in progress) in the title. This is useful when want to get some early feedback.

* PR is supposed to be or near production ready. You should have fixed all styling, adding new tests if possible, and verified the change doesn't break any existing tests. 

* For small changes where the approach seems obvious, you can open a PR with what you believe to be production-ready or near-production-ready code. As you get more experience with how we develop code, you'll find that more PRs will begin falling into this category.

### Commit Messages And Titles of Pull Requests

Overcommit adds some requirements to your commit messages. At Uber, we follow the
[Chris Beams](http://chris.beams.io/posts/git-commit/) guide to writing git
commit messages. Read it, follow it, learn it, love it.

All commit messages are from the titles of your pull requests. So make sure follow the rules when titling them. 
Please don't use very generic titles like "bug fixes". 

All PR titles should start with UPPER case.

Examples:

- [Make sync activity retry multiple times before fetch history from remote](https://github.com/uber/cadence/pull/1379)
- [Enable archival config per domain](https://github.com/uber/cadence/pull/1351)

#### Code Format and Licence headers checking

The project has strict rule about Golang format. You have to run 
```bash
make fmt
```
to re-format your code. Otherwise the CI(buildkite) test will fail.

Also, this project is Open Source Software, and requires a header at the beginning of
all new source files you are adding. To verify that all files contain the header execute:

```bash
make copyright
```

### Code review
We take code reviews very seriously at Cadence. Please don't be deterred if you feel like you've received some hefty feedback. That's completely normal and expected—and, if you're an external contributor, we very much appreciate your contribution!

In this repository in particular, merging a PR means accepting responsibility for maintaining that code for, quite possibly, the lifetime of Cadence. To take on that reponsibility, we need to ensure that meets our strict standards for production-ready code.

No one is expected to write perfect code on the first try. That's why we have code reviews in the first place!

Also, don't be embarrassed when your review points out syntax errors, stray whitespace, typos, and missing docstrings! That's why we have reviews. These properties are meant to guide you in your final scan.

### Addressing feedback
If someone leaves line comments on your PR without leaving a top-level "looks good to me" (LGTM), it means they feel you should address their line comments before merging. 

You should respond to all unresolved comments whenever you push a new revision or before you merge.

Also, as you gain confidence in Go, you'll find that some of the nitpicky style feedback you get does not make for obviously better code. Don't be afraid to stick to your guns and push back. Much of coding style is subjective. 

### Merging
External contributors: you don't need to worry about this section. We'll merge your PR as soon as you've addressed all review feedback(you will get at least one approval) and pipeline runs are all successful(meaning all tests are passing).

