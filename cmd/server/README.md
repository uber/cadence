This is the primary "kitchen sink" server build, which includes all first-party optional plugins, and is used as our Docker-provided binary.

For the most part, this means that day-to-day library upgrades should:
1. Update the main `go.mod`'s dependencies
2. `make tidy`

New submodules we want to include by default should:
1. Add a replace in this `go.mod` like others
2. Import and register it somehow
3. `make tidy`

And if you have problems tidying:
1. Copy/paste all included modules into this `go.mod` (`/go.mod` + `/common/archiver/gcloud/go.mod` currently)
2. `go mod tidy` this submodule and it will probably be correct, `make build` to make sure
3. Commit to save your results!
4. `make tidy` to make sure it's stable
