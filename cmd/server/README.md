This is the primary "kitchen sink" server build, which includes all first-party optional plugins, and is used as our Docker-provided binary.

CI ensures this builds, but note that dependency upgrades in the "main" module may have to be manually repeated here, as they will otherwise trail by one SHA (because that's the only released SHA that can be referenced in the go.mod).

For the most part, this just means:
1. Update the main `go.mod`'s dependencies
2. Copy/paste all included modules into this `go.mod` (/go.mod + `/common/archiver/gcloud/go.mod` currently)
3. `go mod tidy` this submodule and it will probably be correct

Tooling to streamline this is in the next commit, as it cannot be run in this commit due to Go Modules' limitations.
