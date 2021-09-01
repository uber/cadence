# Periodical feature health check workflow tools(aka Canary)

## Build & Run the canary 

In the project root, build cadence canary binary:
   ```
   make cadence-canary
   ```

Then start canary worker & workflow:
   ```
   ./cadence-canary start
   ```
By default, it will load [the configuration in `config/canary/development.yaml`](../config/canary/development.yaml). 
Run `./cadence-canary -h` for details to understand the start options of how to change the loading directory if needed. 

This binary will start the worker, and a cron workflow which trigger every 30s if not exists.

The cron workflow will kick off a bunch of childWOrkflows for all the feature to verify that Cadence server is operating correctly. 

See the list of worklfows in this directory for this canary project. 