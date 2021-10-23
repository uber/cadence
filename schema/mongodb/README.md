What
----
This directory contains the mongodb schema for every database that cadence owns. The directory structure is as follows


```
./schema
   - cadence/               -- Contains schema for default data models
        - schema.json       -- Contains the latest & greatest snapshot of the schema for the keyspace
        - schema.go         -- Contains the collection schema in Golang structs -- because MongoDB collection is shemaless.  
        - versioned
             - v0.1/
             - v0.2/        -- One directory per schema version change
             - v1.0/
                - manifest.json    -- json file describing the change
                - changes.json     -- changes in this version, only [create collection/index/documents] commands are allowed
```

## MongoDB JSON schema format
Below is an example of a schema JSON file containing two commands, for collection/index/documents creation. 
```json
[
  {
    "create": "collectionName"
  },
  {
    "createIndexes": "collectionName",
    "indexes": [
      {
        "key": {
          "fieldNameA": 1,
          "fieldNameB": -1
        },
        "name": "rowType_version"
      }
    ],
    "writeConcern": { "w": "majority" }
  },
  {
    "insert": "collectionName",
    "documents": [
      {
        "fieldNameA": 1,
        "fieldNameB": 0,
        "fieldNameC": "1234"
      },
      {
        "fieldNameA": 2,
        "fieldNameB": 1,
        "fieldNameC": "12344"
      },
      {
        "fieldNameA": 2,
        "fieldNameB": 2,
        "fieldNameC": "12345"
      }
    ],
    "ordered": false
  }
]
```


How
---

Q: How do I update existing schema ?
* Add your changes to schema.json for snapshot
* Create a new schema version directory under ./schema/<>/versioned/vx.x
  * Add a manifest.json
  * Add your changes in a json file