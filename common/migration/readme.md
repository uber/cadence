## Migration

A package for helping with common operations for migrations.

### Dual-Read

This is a function which allows the calling controller or business logic 
to read the data from either the old system or the new system, or both 
and return the data from either.

Conceptually, dual-read migrations tend to go roughly through the 
following phases: 

1. Add new code and continue on old codepath
2. Start calling new codepath and...
   - Returning the old codepath's result 
   - Checking the data returned from the new path is the same or at least ok(optionally) 
3. Calling the new codepath and returning the new data

This library helps by allowing the code to perform this switch at runtime
and handles some of the annoying parts of this switching, such as ensuring that 
the new codepath is not going to wreck things if it's either too slow, panicing or 
otherwise misbehaving. It's not perfect, there's still ways to break the system with 
poorly behaved code, but ignoring the context timeout panics at least should be handled.

How does it work? The intuition is that for each migration case a new
instance would be created of the migration lib... 

```
// assuming a fictitous user controller:

func NewUserController(p params) {

    // a custom comparison for example's sake
    userComparisonFn := func(
        log log.Logger,
        scope metrics.Scope,
        activeFlow *User,
        activeErr error,
        backgroundFlow *User,
        backgroundErr error) (isEqual bool) {
	
	    if activeFlow.UserID == backgroundFlow.UserID {
	        return true 
	    }
	    
	    // the data isn't the same, write some complaining logs
	    log.Error("The data being returned from the flows isn't equal")
	    
	    return false
	}

    migration := NewDualReaderWithCustomComparisonFn[*User](
		dynamicconfig.UserMigrationRollout, // create a string based dynamic config for this rollout
		func(_ ...dynamicconfig.Filter) bool { return true } // always compare results
		p.log,
		p.scope,
		time.Second * 15, // hard cutoff for any background shadow requests in case the go longer than the ctx
		userComparisonFn)

    return Controller{
        migration: migration
        ... 
    }
}

```

... and for the new and old codepaths, they should be able to fit in a function

```
    
    func (c *controller) GetUserController (ctx context.Context, userID string) (*User, error) {

        oldFlow := func(ctx context.Context) (*getUserResponse, error){
            // .. do whatever is necessary to setup the params
            c.getLegacyUserFlow(ctx, requestParams)
        }
        
        newFlow := func(ctx context.Context) (*getUserResponse, error){
            // create request...
            c.getWebscaleUsers(ctx, requestParams, filters)
        }
        
        constraints := migration.Contraints{
            Operation: "get-user-flow",
            Domain: "some-test-domain",
        }

        user, err := c.migration.ReadAndReturnActive(ctx, constraints, oldFlow, newFlow)
        
        // so the user/err here is whatever's in the 'active' flow. That's controlled
        // by the dynamic config mentioned earlier.
        return user, err
}
```

#### Rollout sequence 

With the rollout dynamic config, the following values being returned will controll the behaviour
of the library, with a hopefully fairly self-evident behaviour. By default it will only call the old flow.

Expected rollout states:
- `only-call-old`
- `call-both-return-old`
- `call-both-return-new`
- `only-call-new`

Returning these incorrectly or with random strings will result in it only calling the old flow.

#### Nomenclature: 'old' and 'new', 'active' and 'background'

The 'old' flow is the original state, which is the default starting point and expected to be the 
safest state, with known good data. The 'new' flow is whatever is being migrated to. 

The 'active' flow is whatever flow is presently expected to return data. Initially it'll be the old 
flow, but once the library's switched to `call-both-return-new` the 'active' flow is the new flow and 
the old flow is pushed to the background as a reverse shadow.

#### Comparision and Default Comparison

It's hard to generalize a good way to compare if data between the two flows is valid or correct, so 
it's likely that most use-cases would call for passing in a custom function which is aware of the context
of the data and can make sensible tradeoff decisions about comparison, metrics and logging information etc 

That said, the basic problem of comparing raw structs coming back from various flows is fairly 
straightforward, so a constructor with `NewDualReaderWithDiffComparison` will provide a default 
go.cmp reflection based comparison library whose intent is to ensure equality in responses between the 
two calls.

This not at all performant and a poor choice at any significant throughput or for large return values, 
so use it carefully.