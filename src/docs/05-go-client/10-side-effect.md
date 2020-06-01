---
layout: default
title: Side effect
permalink: /docs/go-client/side-effect
---

# Side effect

`workflow.SideEffect` is useful for short, nondeterministic code snippets, such as getting a random
value or generating a UUID. It executes the provided function once and records its result into the
:workflow: history. `workflow.SideEffect` does not re-execute upon replay, but instead returns the
recorded result. It can be seen as an "inline" :activity:. Something to note about `workflow.SideEffect`
is that, unlike the Cadence guarantee of at-most-once execution for :activity:activities:, there is no such
guarantee with `workflow.SideEffect`. Under certain failure conditions, `workflow.SideEffect` can
end up executing a function more than once.

The only way to fail `SideEffect` is to panic, which causes a :decision_task: failure. After the
timeout, Cadence reschedules and then re-executes the :decision_task:, giving `SideEffect` another chance
to succeed. Do not return any data from `SideEffect` other than through its recorded return value.

The following sample demonstrates how to use `SideEffect`:

```go
encodedRandom := SideEffect(func(ctx cadence.Context) interface{} {
    return rand.Intn(100)
})

var random int
encodedRandom.Get(&random)
if random < 50 {
    ...
} else {
    ...
}
```
