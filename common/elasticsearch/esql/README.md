# ESQL: Translate SQL to Elasticsearch DSL
[![Build Status](https://travis-ci.org/jysui123/esql.svg?branch=master)](https://travis-ci.org/jysui123/esql) [![codecov](https://codecov.io/gh/jysui123/esql/branch/master/graph/badge.svg)](https://codecov.io/gh/jysui123/esql)

Use SQL to query Elasticsearch. ES V6 compatible.

## Supported features
- [x] =, !=, <, >, <=, >=, <>, ()
- [x] AND, OR, NOT
- [x] LIKE, IN, REGEX, IS NULL, BETWEEN
- [x] LIMIT, SIZE, DISTINCT
- [x] COUNT, COUNT(DISTINCT)
- [x] AVG, MAX, MIN, SUM
- [x] GROUP BY, ORDER BY
- [x] HAVING
- [x] column name selection filtering and replacing
- [x] pagination (search after)
- [ ] JOIN
- [ ] nested queries
- [ ] arithmetics and functions
- [ ] comparison between colNames (e.g. colA < colB)


## Usage
Please refer to code and comments in `esql.go`. `esql.go` contains all the apis that an outside user needs. Below shows a simple usage example:
~~~~go
sql := `SELECT COUNT(*), MAX(colA) FROM myTable WHERE colB < 10 GROUP BY colC HAVING COUNT(*) > 20`
e := NewESql()
dsl, _, err := e.ConvertPretty(sql)    // convert sql to dsl
if err == nil {
    fmt.Println(dsl)
}
~~~~
ESQL support custom colName replacement policy and query target processing policy. It can be useful for time queries. User can register functions to let esql to automatically replace certain colNames and convert query target. Below shows an example.
~~~~go
sql := "SELECT colA FROM myTable WHERE colB < 10 AND dateTime = '2015-01-01T02:59:59Z'"
domainID := "CadenceSampleDomain"
// custom policy that change colName like "col.." to "myCol.."
func myFilter1(colName string) bool {
    return strings.HasPrefix(colName, "col")
}
func myReplace(colName string) (string, error) {
    return "myCol"+colName[3:], nil
}
// custom policy that convert formatted time string to unix nano
func myFilter2(colName string) bool {
    return strings.Contains(colName, "Time") || strings.Contains(colName, "time")
}
func myProcess(timeStr string) (string, error) {
    // convert formatted time string to unix nano integer
    parsedTime, _ := time.Parse(defaultDateTimeFormat, timeStr)
    return fmt.Sprintf("%v", parsedTime.UnixNano()), nil
}
// with the 2 policies , converted dsl is equivalent to
// "SELECT myColA FROM myTable WHERE myColB < 10 AND dateTime = '1561678568048000000'
// in which the time is in unix nano format
e := NewESql()
e.SetReplace(myFilter1, myReplace)     // set up filtering policy
e.SetProcess(myFilter2, myProcess)     // set up process policy
dsl, _, err := e.ConvertPretty(sql)    // convert sql to dsl
if err == nil {
    fmt.Println(dsl)
}
~~~~
For Cadence usage, refer to [this](https://github.com/jysui123/esql/blob/master/cadenceDevReadme.md) link.


## Testing
We are using elasticsearch's SQL translate API as a reference in testing. Testing contains 3 basic steps:
- using elasticsearch's SQL translate API to translate sql to dsl
- using our library to convert sql to dsl
- query local elasticsearch server with both dsls, check the results are identical

However, since ES's SQL api is still experimental, there are many features not supported well. For such queries, testing is mannual.

Features not covered yet:
- `LIKE`, `REGEXP` keyword: ES V6.5's sql api does not support regex search but only wildcard (only support shell wildcard `%` and `_`)
- some aggregations are tested by manual check since ES's sql api does not support them well

To run test locally:
- download elasticsearch v6.5 (optional: kibana v6.5) and unzip them
- run `chmod u+x start_service.sh test.sh`
- run `./start_service.sh <elasticsearch_path> <kibana_path>` to start a local elasticsearch server (by default, elasticsearch listens port 9200, kibana listens port 5600)
- run `python gen_test_date.py -dmi 1 1000 20` to insert 1000 documents to the local es
- run `./test.sh TestSQL` to run all the test cases in `/testcases/sqls.txt`
- generated dsls are stored in `dslsPretty.txt` for reference

To customize test cases:
- modify `testcases/sqls.txt`
- run `python gen_test_date.py -h` for guides on how to insert custom data into your lcoal es
- invalid query test cases are in `testcases/sqlsInvalid.txt`


## ES V2.x vs ES V6.5
|Item|ES V2.x|ES v6.5|
|:-:|:-:|:-:|
|missing check|{"missing": {"field": "xxx"}}|{"must_not": {"exist": {"field": "xxx"}}}|
|group by multiple columns|nested "aggs" field|"composite" flattened grouping|


## Attentions
- We assume all the text are store in type `keyword`, i.e., ES does not break the text into separated words. And we use tag `term` to match the whole text rather than a word in the text.
- `must_not` in ES does not share the same logic as `NOT` in sql
- If you want to apply aggregation on some fields, they should be in type `keyword` in ES (set type of a field by put mapping to your table)
- `COUNT(colName)` will include documents w/ null values in that column in ES SQL API, while in esql we exclude null valued documents
- ES SQL API and esql do not support `SELECT DISTINCT`, but we can achieve the same result by `SELECT * FROM table GROUP BY colName`
- ES SQL API does not support `ORDER BY aggregation`, esql support it by applying bucket_sort
- ES SQL API does not support `HAVING aggregation` that not show up in `SELECT`, esql support it
- To use regex query, the column should be `keyword` type, otherwise the regex is applied to all the terms produced by tokenizer from the original text rather than the original text itself


## Current Issue
- parsing issue w/ date datatype in nested field


## Acknowledgement
This project is originated from [elasticsql](https://github.com/cch123/elasticsql). Table below shows the improvement.

|Item|detail|
|:-:|:-:|
|keyword IS|support standard SQL keywords IS NULL, IS NOT NULL for missing check|
|keyword NOT|support NOT, convert NOT recursively since elasticsearch's must_not is not the same as boolean operator NOT in sql|
|keyword LIKE|using "wildcard" tag, support SQL wildcard '%' and '_'|
|keyword REGEX|using "regexp" tag, support standard regular expression syntax|
|keyword GROUP BY|using "composite" tag to flatten multiple grouping|
|keyword ORDER BY|using "bucket_sort" to support order by aggregation functions|
|keyword HAVING|using "bucket_selector" and painless scripting language to support HAVING|
|aggregations|allow introducing aggregation functions from all HAVING, SELECT, ORDER BY|
|column name filtering|allow user pass an white list, when the sql query tries to select column out side white list, refuse the converting|
|column name replacing|allow user pass an function as initializing parameter, the matched column name will be replaced upon the policy|
|query value replacing|allow user pass an function as initializing parameter, query value will be processed by such function if the column name matched in filter function|
|pagination|also return the sorting fields for future search after usage|
|optimization|using "filter" tag rather than "must" tag to avoid scoring analysis and save time|
|optimization|no redundant {"bool": {"filter": xxx}} wrapped|all queries wrapped by {"bool": {"filter": xxx}}|
|optimization|does not return document contents in aggregation query|
|optimization|only return fields user specifies after SELECT|
