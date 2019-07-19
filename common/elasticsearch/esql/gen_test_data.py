#!/usr/bin/python

import requests
import json
import random
import sys

url = 'http://localhost:9200/'
indexingRoute = '/_mapping/_doc'
postDataRoute = '/_doc'

headers = {"Content-type": "application/json"}
# type keyword is often used for aggregation and sorting. It can only searched by its exact value
# type text is often used for text analysis, e.g. search a word in the field
schema0 = {
  "order": 0,
  "index_patterns": [
    "cadence-visibility-*"
  ],
  "settings": {
    "index": {
      "number_of_shards": "5",
      "number_of_replicas": "0"
    }
  },
  "mappings": {
    "_doc": {
      "dynamic": "false",
      "properties": {
        "DomainID": {
          "type": "keyword"
        },
        "WorkflowID": {
          "type": "keyword"
        },
        "RunID": {
          "type": "keyword"
        },
        "WorkflowType": {
          "type": "keyword"
        },
        "StartTime": {
          "type": "long"
        },
        "ExecutionTime": {
          "type": "long"
        },
        "CloseTime": {
          "type": "long"
        },
        "CloseStatus": {
          "type": "integer"
        },
        "HistoryLength": {
          "type": "integer"
        },
        "KafkaKey": {
          "type": "keyword"
        },
        "Attr": {
          "properties": {
            "CustomStringField":  { "type": "text" },
            "CustomKeywordField": { "type": "keyword"},
            "CustomIntField": { "type": "long"},
            "CustomBoolField": { "type": "boolean"},
            "CustomDoubleField": { "type": "double"},
            "CustomDatetimeField": {"type": "date",
            "format": "strict_date_optional_time||yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"}
          }
        }
      }
    }
  },
  "aliases": {}
}

schema1 = {
    "mappings": {
        "_doc": {
        "dynamic": "false",
            "properties": {
                "colA": {"type": "keyword"},
                "colB": {"type": "keyword"},
                "colC": {"type": "keyword"},
                "colD": {"type": "long"},
                "colE": {"type": "double"},
                "ExecutionTime": {"type": "long"},
            }
        }
    }
}

def genRandStr(letters="abc", len=3):
    return "".join(random.choice(letters) for i in range(len))

def padZero(s, length=2):
    return '0'*(length-len(s)) + s

def genDate(precision="d", startYear=2010):
    dateStr = ""
    if precision in ["y", "M", "d", "h", "m", "s", "ms"]:
        dateStr = str(startYear + random.randint(0, 5))
    if precision in ["M", "d", "h", "m", "s", "ms"]:
        dateStr = dateStr + "-" + padZero(str(random.randint(1, 12)))
    if precision in ["d", "h", "m", "s", "ms"]:
        dateStr = dateStr + "-" + padZero(str(random.randint(1, 28)))
    if precision in ["h", "m", "s", "ms"]:
        dateStr = dateStr + "T" + padZero(str(random.randint(0, 23)))
    if precision in ["m", "s", "ms"]:
        dateStr = dateStr + ":" + padZero(str(random.randint(0, 59)))
    if precision in ["s", "ms"]:
        dateStr = dateStr + ":" + padZero(str(random.randint(0, 59)))
        if not precision == "ms":
          dateStr = dateStr + "Z"
    if precision in ["ms"]:
        dateStr = dateStr + "." + padZero(str(random.randint(0, 999)), 3)
    return dateStr


def genPayload(tableName, missing=20):
    payload = {}
    if tableName == "test0":
        payload['DomainID'] = genRandStr("012", 1)
        payload['RunID'] = genRandStr('abcdefghijklmnopqrstuvwxyz', 8)
        payload['WorkflowID'] = genRandStr('abcdefghijklmnopqrstuvwxyz', 8)
        payload['WorkflowType'] = "main.workflow"+genRandStr('012', 1)
        payload['StartTime'] = random.randint(-500, 1200000000000000000)
        payload['ExecutionTime'] = random.randint(0, 200000000000000000)
        payload['CloseTime'] = payload['StartTime'] + payload['ExecutionTime']
        payload['CloseStatus'] = random.randint(0, 2)
        payload['HistoryLength'] = random.randint(1, 100)
        payload['KafkaKey'] = genRandStr('abcdefghijklmnopqrstuvwxyz', 8)

        attr = {}
        attr['CustomStringField'] = genRandStr("abc", 3) + ' ' + genRandStr("abc", 3)
        attr['CustomKeywordField'] = genRandStr("ab", 2)
        attr['CustomIntField'] = random.randint(-500, 2000)
        attr['CustomBoolField'] = False
        attr['CustomDoubleField'] = random.uniform(0, 20)
        attr['CustomDatetimeField'] = genDate('s')
        payload['Attr'] = attr

    elif tableName == "test1":
        payload['colA'] = genRandStr()
        payload['colB'] = genRandStr("ab", 2)
        payload['colC'] = payload['colA'] + " " + genRandStr() + " " + genRandStr()
        payload['colD'] = random.randint(0, 20)
        payload['colE'] = random.uniform(0, 20)
        payload['ExecutionTime'] = random.randint(-500, 2000)

    finalPayload = {}
    for k, v in payload.iteritems():
        if random.randint(1, 100) > missingPercent:
            finalPayload[k] = v
    return finalPayload


def insertData(tableName, nRows, missingPercent):
    for i in range(nRows):
        payload = genPayload(tableName, missingPercent)
        resp = requests.post(url+tableName + postDataRoute, data=json.dumps(payload), headers=headers)
        if resp == None or resp.status_code != 201:
            print("cannot insert data: {}: {}\n".format(resp.status_code, requests.status_codes._codes[resp.status_code]))
            print(i, json.dumps(payload))
            exit(1)
    print("successfully insert {} documents (rows)".format(nRows))

def putMapping(tableName):
    sch = schema0 if tableName == "test0" else schema1
    resp = requests.put(url+tableName, data=json.dumps(sch), headers=headers)
    if resp == None or resp.status_code != 200:
        print("cannot put mapping: {}: {}".format(resp.status_code, requests.status_codes._codes[resp.status_code]))
        exit(1)
    print("successfully put mapping")

def deleteIndex(tableName):
    resp = requests.delete(url+tableName)
    if resp == None or resp.status_code not in [200, 202, 204]:
        print("cannot delete index")
    print("successfully delete index")


# usage: python gen_test_data.py -dmi <index id> <number of rows> <missing rate>
nRows = 200
missingPercent = 0
tableNum = 1
if len(sys.argv) > 5:
    print ("too many arguments")
    exit(1)

if sys.argv[1] == "-h":
    print ("usage: python gen_test_data.py -dmi <index id> <number of rows> <missing rate>")
    print ("       -d: delete index")
    print ("       -i: generate and insert documents")
    print ("       -m: put index")
    print ("       index id: 0 or 1 indicates different schemas, ordinary test use 1")
    print ("       number of rows: how many documents(rows) to generate")
    print ("       missing rate: the rate that a field is null, mainly used for missing check (IS NULL) test")
    exit(0)

if len(sys.argv) > 2:
    tableNum = int(sys.argv[2])

if len(sys.argv) > 3:
    nRows = int(sys.argv[3])

if len(sys.argv) > 4:
    missingPercent = int(sys.argv[4])

tableName = 'test' + str(tableNum)
for i in range(len(sys.argv[1])):
    if i == 0:
        if sys.argv[1][i] != '-':
            print("invalid argument")
            exit(1)
    else:
        if sys.argv[1][i] == 'i':
            insertData(tableName, nRows, missingPercent)
        elif sys.argv[1][i] == 'd':
            deleteIndex(tableName)
        elif sys.argv[1][i] == 'm':
            putMapping(tableName)
        else:
            print("invalid argument, allowed: -idm")
            exit(1)

