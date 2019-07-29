# Command Line Interface

The Cadence CLI is a command-line tool you can use to perform various tasks on a Cadence server. It can perform 
domain operations such as register, update, and describe as well as workflow operations like start 
workflow, show workflow history, and signal workflow.

## Using the CLI

The Cadence CLI can be used directly from the Docker Hub image *ubercadence/cli* or by building the CLI tool
locally. 

Example of using the docker image to describe a domain:
```
docker run --rm ubercadence/cli:master --domain samples-domain domain describe
```

You could use docker image `ubercadence/cli`, by replacing all the following `./cadence ...` with `docker run --rm ubercadence/cli:master --address <HOST_DNS>:7933 ...` . 

You may get "connection refused" error for 18.03 onwards versions of docker. You have to use host.docker.internal. For more info check https://docs.docker.com/docker-for-mac/networking/#use-cases-and-workarounds  
```
docker run --rm ubercadence/cli:master --address host.docker.internal:7933 --domain samples-domain domain describe
```


To build the CLI tool locally, clone the [Cadence server repo](https://github.com/uber/cadence) and run
`make bins`. This produces an executable called `cadence`. With a local build, the same command to 
describe a domain would look like this:
```
./cadence --domain samples-domain domain describe
```

The example commands below will use `./cadence` for brevity.

## Environment variables

Setting environment variables for repeated parameters can shorten the CLI commands.

- **CADENCE_CLI_ADDRESS** - host:port for Cadence frontend service, the default is for the local server
- **CADENCE_CLI_DOMAIN** - default workflow domain, so you don't need to specify `--domain`

## Quick Start
Run `./cadence` for help on top level commands and global options   
Run `./cadence domain` for help on domain operations  
Run `./cadence workflow` for help on workflow operations  
Run `./cadence tasklist` for help on tasklist operations  
(`./cadence help`, `./cadence help [domain|workflow]` will also print help messages)

**Note:** make sure you have a Cadence server running before using CLI 

### Domain operation examples 
- Register a new domain named "samples-domain":  
```
./cadence --domain samples-domain domain register --global_domain false
# OR using short alias  
./cadence --do samples-domain d re --gd false
```   
- View "samples-domain" details:   
```
./cadence --domain samples-domain domain describe  
```

### Workflow operation examples
The following examples assume the CADENCE_CLI_DOMAIN environment variable is set.

#### Run workflow
Start a workflow and see its progress. This command doesn't finish until workflow completes.
```
./cadence workflow run --tl helloWorldGroup --wt main.Workflow --et 60 -i '"cadence"'

# view help messages for workflow run
./cadence workflow run -h
``` 
Brief explanation:  
To run a workflow, the user must specify the following:
1. Tasklist name (--tl)
2. Workflow type (--wt)
3. Execution start to close timeout in seconds (--et)
4. Input in JSON format (--i) (optional) 

This example uses [this cadence-samples workflow](https://github.com/samarabbas/cadence-samples/blob/master/cmd/samples/recipes/helloworld/helloworld_workflow.go) 
and takes a string as input with the `-i '"cadence"'` parameter. Single quotes (`''`) are used to wrap input as JSON. 

**Note:** You need to start the worker so that the workflow can make progress.  
(Run `make && ./bin/helloworld -m worker` in cadence-samples to start the worker)

#### Show running workers of a tasklist
```
./cadence tasklist desc --tl helloWorldGroup
```

#### Start workflow
```
./cadence workflow start --tl helloWorldGroup --wt main.Workflow --et 60 -i '"cadence"'

# view help messages for workflow start
./cadence workflow start -h

# for a workflow with multiple inputs, separate each json with space/newline like
./cadence workflow start --tl helloWorldGroup --wt main.WorkflowWith3Args --et 60 -i '"your_input_string" 123 {"Name":"my-string", "Age":12345}'
```
The workflow `start` command is similar to the `run` command, but immediately returns the workflow_id and 
run_id after starting the workflow. Use the `show` command to view the workflow's history/progress.  

##### Reuse the same workflow id when starting/running a workflow

Use option `--workflowidreusepolicy` or `--wrp` to configure the workflow id reuse policy.  
**Option 0 AllowDuplicateFailedOnly:** Allow starting a workflow execution using the same workflow ID when a workflow with the same workflow ID is not already running and the last execution close state is one of *[terminated, cancelled, timedout, failed]*.  
**Option 1 AllowDuplicate:** Allow starting a workflow execution using the same workflow ID when a workflow with the same workflow ID is not already running.  
**Option 2 RejectDuplicate:** Do not allow starting a workflow execution using the same workflow ID as a previous workflow.  
```
# use AllowDuplicateFailedOnly option to start a workflow
./cadence workflow start --tl helloWorldGroup --wt main.Workflow --et 60 -i '"cadence"' --wid "<duplicated workflow id>" --wrp 0

# use AllowDuplicate option to run a workflow
./cadence workflow run --tl helloWorldGroup --wt main.Workflow --et 60 -i '"cadence"' --wid "<duplicated workflow id>" --wrp 1
```

##### Start a workflow with a memo
Memos are immutable key/value pairs that can be attached to a workflow run when starting the workflow. These are 
visible when listing workflows. More information on memos can be found 
[here]({{ '/docs/02_key_features#memos' | relative_url }}).
```
cadence wf start -tl helloWorldGroup -wt main.Workflow -et 60 -i '"cadence"' -memo_key ‘“Service” “Env” “Instance”’ -memo ‘“serverName1” “test” 5’
```

#### Show workflow history
```
./cadence workflow show -w 3ea6b242-b23c-4279-bb13-f215661b4717 -r 866ae14c-88cf-4f1e-980f-571e031d71b0
# a shortcut of this is (without -w -r flag)
./cadence workflow showid 3ea6b242-b23c-4279-bb13-f215661b4717 866ae14c-88cf-4f1e-980f-571e031d71b0

# if run_id is not provided, it will show the latest run history of that workflow_id
./cadence workflow show -w 3ea6b242-b23c-4279-bb13-f215661b4717
# a shortcut of this is
./cadence workflow showid 3ea6b242-b23c-4279-bb13-f215661b4717
```

#### Show workflow execution information
```
./cadence workflow describe -w 3ea6b242-b23c-4279-bb13-f215661b4717 -r 866ae14c-88cf-4f1e-980f-571e031d71b0
# a shortcut of this is (without -w -r flag)
./cadence workflow describeid 3ea6b242-b23c-4279-bb13-f215661b4717 866ae14c-88cf-4f1e-980f-571e031d71b0

# if run_id is not provided, it will show the latest workflow execution of that workflow_id
./cadence workflow describe -w 3ea6b242-b23c-4279-bb13-f215661b4717
# a shortcut of this is
./cadence workflow describeid 3ea6b242-b23c-4279-bb13-f215661b4717
```

#### List closed or open workflow executions
```
./cadence workflow list

# default will only show one page, to view more items, use --more flag
./cadence workflow list -m
```

#### Query workflow execution
```
# use custom query type
./cadence workflow query -w <wid> -r <rid> --qt <query-type>

# use build-in query type "__stack_trace" which is supported by Cadence client library
./cadence workflow query -w <wid> -r <rid> --qt __stack_trace
# a shortcut to query using __stack_trace is (without --qt flag)
./cadence workflow stack -w <wid> -r <rid> 
```

#### Signal, cancel, terminate workflow
```
# signal
./cadence workflow signal -w <wid> -r <rid> -n <signal-name> -i '"signal-value"'

# cancel
./cadence workflow cancel -w <wid> -r <rid>

# terminate
./cadence workflow terminate -w <wid> -r <rid> --reason 
```
Terminating a running workflow execution will record a WorkflowExecutionTerminated event as the closing event in the history. No more decision tasks will be scheduled for a terminated workflow execution.  
Canceling a running workflow execution will record a WorkflowExecutionCancelRequested event in the history, and a new decision task will be scheduled. The workflow has a chance to do some clean up work after cancellation.

#### Restart, reset workflow
The Reset command allows resetting a workflow to a particular point and continue running from there.
There are a lot of use cases:
- Rerun a failed workflow from the beginning with the same start parameters.
- Rerun a failed workflow from the failing point without losing the achieved progress(history).
- After deploying new code, reset an open workflow to let the workflow run to different flows.

You can reset to some predefined event types:
```
./cadence workflow reset -w <wid> -r <rid> --reset_type <reset_type> --reason "some_reason"
```

- FirstDecisionCompleted: reset to the beginning of the history.
- LastDecisionCompleted: reset to the end of the history.
- LastContinuedAsNew: reset to the end of the history for the previous run.

If you are familiar with the Cadence history event, You can also reset to any decision finish event by using:
```
./cadence workflow reset -w <wid> -r <rid> --event_id <decision_finish_event_id> --reason "some_reason"
```
Some things to note:
- When reset, a new run will be kicked off with the same workflowID. But if there is a running execution for the workflow(workflowID), the current run will be terminated.
- decision_finish_event_id is the ID of events of the type: DecisionTaskComplete/DecisionTaskFailed/DecisionTaskTimeout.
- To restart a workflow from the beginning, reset to the first decision task finish event.

To reset multiple workflows, you can use batch reset command:
```
./cadence workflow reset-batch --input_file <file_of_workflows_to_reset> --reset_type <reset_type> --reason "some_reason"
```
#### Recovery from bad deployment -- auto-reset workflow
If a bad deployment lets a workflow run into a wrong state, you might want to reset the workflow to the point that the bad deployment started to run. But usually it is not easy to find out all the workflows impacted, and every reset point for each workflow. In this case, auto-reset will automatically reset all the workflows given a bad deployment identifier. 

Let's get familar with some concepts. Each deployment will have an identifier, we call it "**Binary Checksum**" as it is usually generated by the md5sum of a binary file. For a workflow, each binary checksum will be associated with an **auto-reset point**, which contains a **runID**, an **eventID**, and the **created_time** that binary/deployment made the first decision for the workflow.

To find out which **binary checksum** of the bad deployment to reset, you should be aware of at least one workflow running into a bad state. Use the describe command with **--reset_points_only** option to show all the reset points:
```
./cadence wf desc -w <WorkflowID>  --reset_points_only
+----------------------------------+--------------------------------+--------------------------------------+---------+
|         BINARY CHECKSUM          |          CREATE TIME           |                RUNID                 | EVENTID |
+----------------------------------+--------------------------------+--------------------------------------+---------+
| c84c5afa552613a83294793f4e664a7f | 2019-05-24 10:01:00.398455019  | 2dd29ab7-2dd8-4668-83e0-89cae261cfb1 |       4 |
| aae748fdc557a3f873adbe1dd066713f | 2019-05-24 11:01:00.067691445  | d42d21b8-2adb-4313-b069-3837d44d6ce6 |       4 |
...
...
```

Then use this command to tell Cadence to auto-reset all workflows impacted by the bad deployment. The command will store the bad binary checksum into domain info and trigger a process to reset all your workflows.
```
./cadence --do <YourDomainName> domain update --add_bad_binary aae748fdc557a3f873adbe1dd066713f  --reason "rollback bad deployment"
```

As you add the bad binary checksum to your domain, Cadence will not dispatch any decision tasks to the bad binary. So make sure that you have rolled back to a good deployment(or roll out new bits with bug fixes). Otherwise your workflow can't make any progress after auto-reset. 
