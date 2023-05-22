Documentation for the Cadence command line interface is located at our [main site](https://cadenceworkflow.io/docs/cli/).

### Build CLI binary locally

To build the CLI tool locally check out the version tag (e.g. `git checkout v0.21.3`) and run `make tools`. 
This produces an executable called cadence.

Run help command with a local build:
````
./cadence --help
````

Command to describe a domain would look like this:
````
./cadence --domain samples-domain domain describe
````