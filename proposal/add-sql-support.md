# Proposal: Adding support for SQL

Author(s): Alex Drinkwater & Ryan Walls

Status: Initial Design

## Summary
Remove cadence-cassandra-tool and use [migrate](https://github.com/golang-migrate/migrate) instead.
Add support for SQL lite.

## Objectives

- Ability to easily add different database types in the future
- Remove cadence-cassandra-tool
- Use a 3rd party library to manage migrations
- Cadence manages migrations on start (keeps cadence version tied to correct schema version)

## Motivation
We are using cadence for a desktop application and Cassandra is over kill 
for our needs with only a single instance of cadence running. In order to 
remove the overhead of running cassandra we need to be able to support other
Databases. For our desktop use case we wish to use SQL lite. While we are at it
we figured it would be best to implement the change in a way that allows more than
just SQL lite to be added.

## Proposal
We propose to use a 3rd party library to manage the migrations of the databases.

The library we are proposing to use is: [migrate](https://github.com/golang-migrate/migrate)

Migrate supports multiple types of [databases](https://github.com/golang-migrate/migrate#databases)

### Implementation Steps

1. Import migrate
2. 

