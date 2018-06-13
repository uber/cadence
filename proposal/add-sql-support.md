# Proposal: Adding support for SQL databases

Author(s): Alex Drinkwater, Ryan Walls

Status: Initial Design

## Summary
Add support for SQL databases in addition to the existing Cassandra support.  

## Objectives

- Easily add different SQL database implementations in the future
- Use a 3rd party library to manage migrations (considering https://github.com/golang-migrate/migrate)
- Have Cadence manage migrations on start (keeps cadence version tied to correct schema version)
- Implement sqlite support
- Make it easy to add other SQL databases, such as Microsoft SQL server  

## Motivation
We are using Cadence for a desktop application and Cassandra uses too much memory
for our needs. Therefore, we would like to be able to support other
databases, including at least one embedded database.  We would also like to use an Amazon RDS supported database in our 
cloud deployment as we are not proficient at running Cassandra operationally yet.    

## Proposal
The proposal is to add support for SQL databases.  We would target sqlite and MS SQL server initially.  
We propose using a 3rd party library to manage the migrations of the databases, such as 
[migrate](https://github.com/golang-migrate/migrate).  Migrate supports multiple types of [databases](https://github.com/golang-migrate/migrate#databases)

### Implementation Summary

1. [Use migrate for existing cassandra schema updates](#Use-migrate-for-existing-cassandra-schema-updates)
2. [Implement SQL lite as a persistence option](#Implement-SQL-lite-as-a-persistence-option)
3. [Implement schema migrations support for SQL lite](#Implement-schema-migrations-support-for-SQL-lite)


#### Use migrate for existing cassandra schema updates 
Remove existing cadence-cassandra-tool and use migrate library within cadence to manage schema.
This will make sure that cadence is always in sync with the underlying data structures.  
We could skip this step if we're okay using two different schema migration strategies.

Note: if schema migration was required to be kept separate from cadence the [command line version](https://github.com/golang-migrate/migrate#cli-usage)
of migrate library could be used.

#### Implement SQL lite as a persistence option
Implement dataInterfaces.go for sqlite (maybe using an ORM like [GORM](https://github.com/jinzhu/gorm))

#### Implement schema migrations support for sqlite
Using schema management tool, like Migrate, setup migration strategy for SQL based databases.  If we use sqlite we will 
need to be very careful to only use sqlite-compatible migrations (See https://www.sqlite.org/lang_altertable.html)