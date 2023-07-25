# Scanner / fixer / etc how-to

For future scanner-extenders/rewriters, this is intended as a rough guide on what to change and how it works.

That said, we should probably rewrite it, so please don't take this as evidence that this is an architecture we want to keep.  For a lot of reasons:
- To allow external extension without modifying this repository, which is currently not possible.
- To allow scan-and-fix in one step, to reduce delay and duplicate operations.  (in large clusters, scanner can take *weeks* to complete)
- There is a large amount of loose stringly-typed casting / binding spread across many files.

Files in this doc are linked by relative paths, which works on GitHub and in most IDEs.
Sadly, links to specific lines and repo-relative `/path/to/file` support is widely varying, so it should be avoided.

# Overall behavior

At the highest level, keep in mind:
- Invariants define things to check and fix, and are defined 1:1 with their Collections but do not share names directly
  - These are what actually check if X is bad, and do the things necessary to fix it
  - Each handles only one source/type for their primary data, e.g. concrete vs current vs timer
  - Invariant Managers are just N invariants that are run in a list.  Note that they are _not_ related to invariant collections.
- Scanner and Fixer can be enabled for per data type (separate workflows) and per Collection (via config only)
- Scanner and Fixer are overall very similar:
  - Consume a data-iterator (from database query or blobstore)
  - Run their invariant manager on that data (check or fix)
  - Push detailed results to a blob store
- Concrete's and Current's invariants/workflows/etc are often side by side and structured similarly
  - Timer's are less similar and sometimes in a slightly different location, e.g. [scanner/timers](timers) vs [shardscanner](shardscanner)
- _Most_ details covered below can have "scanner" or "fixer" used interchangeably...
  but not all, and the details are not often worth detailing and might change anyway.  Read the code before assuming anything.

## Startup

On workers-service startup, [scanner.go](scanner.go) checks config:
- `ConcreteExecutionsScannerEnabled` (and `ConcreteExecutionFixerEnabled`)
- `CurrentExecutionsScannerEnabled` (and fixer)
- `TimersScannerEnabled` (and fixer)

For each thing that is enabled:
- Their workflow is started via the RPC client
- The tasklist is returned, so that worker is started
- The worker's shared background context is wrapped to add each thing's
  dependencies _under its workflow name as a context key_, in [shardscanner/types.go](shardscanner/types.go), e.g.:
  ```go
  ctx = shardscanner.NewFixerContext( // also scanner
      ctx,
      config.FixerWFTypeName,
      shardscanner.NewShardFixerContext(s.context.resource, config),
  )
  ```
- These are later pulled out in the activities, via `GetFixerContext(ctx)` (and scanner).

## Workflows

The workflows are fairly simple and both scanner and fixer are almost identical.
Code is almost completely defined in:
- [shardscanner/fixer_workflow.go](shardscanner/fixer_workflow.go)
- [shardscanner/scanner_workflow.go](shardscanner/scanner_workflow.go)

Essentially they just:
- Run a config-reading activity to read the dynamic config and save it into history.
- Fixer also _queries_ the most-recent scanner activity to get the range of keys to process.
  - I'm not sure why this queries rather than reading the result.
- Start up N "run Scanner/Fixer" activity instances (split into shard-range batches, running with configured concurrency).
  - Results are aggregated and reported at the end.

Fixer does a small (unnecessary AFAICT) indirection out to the individual definition files,
e.g. [executions/concrete_executions.go](executions/concrete_execution.go) defines a `ConcreteFixerWorkflow`
which does nothing but create a `FixerWorkflow` (in [shardscanner/fixer_workflow.go](shardscanner/fixer_workflow.go))
and call `Start` on it... which could have been avoided by getting the `GetInfo(ctx).WorkflowTypeName` rather than
using the indirection to pass in a hardcoded workflow name.  
But ultimately all behavior is still in the [shardscanner/fixer_workflow.go](shardscanner/fixer_workflow.go) and the
Fixer it creates rather than elsewhere.

## Config-reading activity

One important part of this config is the `CustomScannerConfig`, which controls which invariants are enabled.
This is essentially a serialized set created from dynamic config:
```go
func ConcreteExecutionConfig(ctx shardscanner.Context) shardscanner.CustomScannerConfig {
	res := shardscanner.CustomScannerConfig{}
	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.ConcreteExecutionsScannerInvariantCollectionHistory)() {
		res[invariant.CollectionHistory.String()] = strconv.FormatBool(true)
	}
	...
	return res
}
```
Which is later turned into the enabled invariants:
```go
func ParseCollections(params shardscanner.CustomScannerConfig) []invariant.Collection {
	var collections []invariant.Collection
	for k, v := range params {
		c, e := invariant.CollectionString(k)
		if e != nil { continue }               // must be a valid invariant name
		enabled, err := strconv.ParseBool(v)
		if enabled && err == nil {             // must be "true"
			collections = append(collections, c)
		}
	}
	return collections
}
```
The values are read, but they are only ever set to "true" when the values are true.

This set of invariants later becomes the invariant manager that is used:
```go
func ScannerManager(...) invariant.Manager {
	collections := ParseCollections(params.ScannerConfig)
	var ivs []invariant.Invariant
	for _, fn := range ConcreteExecutionType.ToInvariants(collections, zap.NewNop()) {
		ivs = append(ivs, fn(pr, domainCache))
	}
	return invariant.NewInvariantManager(ivs)
}
```

This allows you to enable/disable individual invariants with dynamic config like `worker.executionsScannerInvariantCollectionStale`.  
These can only take effect if their hosting scanner/fixer is also enabled, e.g. `worker.concreteExecutionFixerEnabled`.

## Activities and their context

At a very high level, both scanner's and fixer's main activities are almost identical:
- Plucks out its activity-context by workflow-name with all its implementations / config / etc.
- Loops over all shards it was told are in its batch, and one by one:
  - Creates a Scanner or Fixer instance for that shard, and tells it to run.

They're even defined in the same file: [shardscanner/activities.go](shardscanner/activities.go)

All scanner/fixer activities use the same base context-passed objects, but add
custom types / callbacks / etc to get their work done.  Mostly this is handled via
the scanner or fixer `Hooks()`, which come from config, which come from the scanner type.

E.g. concretely, for concrete executions, see [executions/concrete_execution.go](executions/concrete_execution.go).
In there, you can see lines like this:
```go
func ConcreteExecutionScannerConfig(dc *dynamicconfig.Collection) *shardscanner.ScannerConfig {
	return &shardscanner.ScannerConfig{
		ScannerWFTypeName: ConcreteExecutionsScannerWFTypeName, // also used as context key
		FixerWFTypeName:   ConcreteExecutionsFixerWFTypeName,   // also used as context key
		DynamicParams:     shardscanner.DynamicParams{...},     // contains enabled flags and other limits
		DynamicCollection: dc,
		ScannerHooks:      ConcreteExecutionHooks,       // constructor-func for this impl
		FixerHooks:        ConcreteExecutionFixerHooks,  // constructor-func for this impl
		...
	}
}
```
There's quite a large chunk of complicated code reuse, value-driven behavior rather than type-driven, and back and forth
between changing based on arguments or hardcoded class-method-like lookups underneath all this, but the hooks themselves
are relatively readable after only a couple layers of navigation.

They broadly summarize as:
- Config has hooks and dynamic config, some of which defines invariants.
- Hooks have a "manager" which basically collects invariants.
- Hooks have an "iterator" which yields a page of data each time it is called (the shared iterator handles forwarding paging tokens).
- Some of all of the above behavior is achieved by using methods on the iota that defines the invariant
  (which is often selected from serialized data by its string name).

Some of these fields and stuff from the high-level shared `Resources` object are combined
to build the `ShardScanner` or `ShardFixer` per shard, which is the thing that drives the "read -> process -> write" loops
at the core of these processes:
```go
scanner := NewScanner(
    shardID,
    ctx.Hooks.Iterator(activityCtx, pr, params),
    resources.GetBlobstoreClient(),
    params.BlobstoreFlushThreshold,
    ctx.Hooks.Manager(activityCtx, pr, params, resources.GetDomainCache()),
    func() { activity.RecordHeartbeat(activityCtx, heartbeatDetails) },
    scope,
    resources.GetDomainCache(),
)
```

And last but not least, note that:
- Config is saved when the workflow starts, not read from scratch when the activities begin.
- Scanner/fixer are on an every-minute cron schedule, but are allowed to run for ludicrous amounts of time.
- This means any config changes will _very likely_ need to wait for the current run to finish before picking up your changes.
  - You can advance this relatively quickly by terminating the existing workflows, and then restarting a workers-service.
  - Terminating will stop the cron completely, but restarting any instance will start a new cron, which will pick up the new config.

<details><summary>Enabling scanners for local dev</summary>

For local purposes, you can pretty easily enable these via the dynamicconfig
file store [config/dynamicconfig/development.yaml](../../../config/dynamicconfig/development.yaml):
```yaml
# inconsistent key names, unfortunately.  don't guess, copy/paste
worker.concreteExecutionFixerEnabled:
- value: true
  constraints: {}
worker.executionsScannerInvariantCollectionStale:
- value: true
  constraints: {}
worker.executionsScannerEnabled:
- value: true
  constraints: {}
```

</details>

## ShardScanner and ShardFixer implementations

Implemented in:
- [shardscanner/scanner.go](shardscanner/scanner.go)
- [shardscanner/fixer.go](shardscanner/fixer.go)

Once again, these are fairly similar structurally:
- Iterate over their data (from the scan-paginator or the blobstore to-fix paginator)
- Run the invariant manager's check or fix on each item
- Collect aggregated metrics
- Periodically flush details to a blob-store

# Some other files involved that seemed interesting to link

- [/common/reconciliation/invariant/types.go](../../../common/reconciliation/invariant/types.go)
  - `Name` entries name and refer to individual `Invariant` implementations.
  - `Collection` entries define... categories of sorts, which are used elsewhere to configure which `Invariant`s run.
- [/service/worker/scanner/executions/types.go](executions/types.go)
  - `Collection` selections elsewhere are converted to `Invariant`s, based on the scanner type
    - Timer scanners are handled in [scanner/timers/timers.go](timers/timers.go) rather than in here
- [/service/worker/scanner/executions/concrete_execution.go](executions/concrete_execution.go)
  - this and [current_execution.go](executions/current_execution.go) define a `[ExecutionType]ExecutionConfig` func that can override which `Collection`s are run in a scanner.

# Wishlist

- Collections and how config is serialized -> checked against hardcoded collections -> re-serialized repeatedly
  make it impossible to add invariants in external code.
  - There are _many_ ways this could be made both simpler and more flexible.
  - This is also a major reason it's difficult to e.g. add loggers or change APIs.
    They're repeatedly treated identically, then hardcoded, then identical, etc and customizations need to be threaded through all stages.
  - The config-reading activity at the start of the workflows is a prime example of this.
    - It could have just been a list of invariant names, and all this logic could have been handled by `encoding/json` plus a duplicate check in the invariant manager or something.
    - Instead it introduces custom serialization far away from the type, uses iotas and enumer codegen and switch statements which share many logically-incompatible types,
      has a `map[string]string` pulling double duty as a set (it is at least uniquely typed!), doesn't complain about passing incorrect keys (a wrong-type invariant is ignored),
      uses multiple different and disconnected kinds of strings and keys and values to control if an invariant is included or not, and is spread across several files.
- "Don't run the workers if they are not configured" seems... maybe not ideal?
  - Some benefits, some downsides, but definitely more complicated.
  - Maybe just run these workers all the time?  Workflows/etc can be canceled if you want to stop them, the current setup only _pauses_ them (with essentially no timeout).
- "Scan everything then fix everything" has pretty clear downsides, not least of which is potentially-massive latency.
  - Being _able_ to separate them is useful for verification, that should be kept.
- Invariants (some anyway) are used elsewhere, which is part of why they're highly generic.
  - I have not investigated this use of invariants.
  - Might only be related to NDC replication?
  - Oddly, these appear to only use a _single_ invariant field, not an invariant Manager.
- Almost everything that is an iota / integer would be MUCH better off as a typed string or similar.
  - I'm surprised how well goland converts these to known types, but there are many ways it fails.
    Strings are just far, FAR more likely to be immediately understood, serialize/deserialize predictably, refactor safely, etc.
  - The only thing we gain from iota is uniqueness... and that's whitespace sensitive.
