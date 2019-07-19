// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package esql

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/olivere/elastic"
)

var tableName = `test`
var testCases = `testcases/sqls.txt`
var testCasesInvalid = `testcases/sqlsInvalid.txt`
var testCasesInvalidCad = `testcases/sqlsInvalidCad.txt`
var testCasesBenchmarkAll = `testcases/sqlsBm.txt`
var testCasesBenchmarkAgg = `testcases/sqlsBmAgg.txt`
var testCasesBenchmarkCadence = `testcases/sqlsBmCad.txt`
var testCasesCad = `testcases/sqlsCad.txt`
var testDsls = `testcases/dsls.txt`
var testDslsCadence = `testcases/dslsCadence.txt`
var testDslsPretty = `testcases/dslsPretty.txt`
var testDslsPrettyCadence = `testcases/dslsPrettyCadence.txt`
var testBenchmarkDslsEsql = `testcases/dslsBmEsql.txt`
var testBenchmarkDslsElasticsql = `testcases/dslsBmElasticsql.txt`
var testBenchmarkCntEsql = `testcases/dslsBmEsqlCnt.txt`
var testBenchmarkCntElasticsql = `testcases/dslsBmElasticsqlCnt.txt`
var groundTruth = ``
var urlES = "http://localhost:9200"
var url = "http://localhost:9200/test1/_search"
var urlSQL = "http://localhost:9200/_xpack/sql/translate"
var index = "test1"
var notTestedKeywords = []string{"LIMIT", "LIKE", "REGEX"}

func compareRespGroup(respES *elastic.SearchResult, resp *elastic.SearchResult) error {
	var groupES, group map[string]interface{}
	if _, exist := respES.Aggregations["groupby"]; !exist {
		fmt.Printf("\tagg without group by, not covered in test module\n")
		return nil
	}
	respESbyte, _ := respES.Aggregations["groupby"].MarshalJSON()
	err := json.Unmarshal(respESbyte, &groupES)
	if err != nil {
		return err
	}
	respByte, _ := resp.Aggregations["groupby"].MarshalJSON()
	err = json.Unmarshal(respByte, &group)
	if err != nil {
		return err
	}
	bucketCounts := make(map[int]int)
	buckNum := 0
	for _, bucket := range groupES["buckets"].([]interface{}) {
		if b, ok := bucket.(map[string]interface{}); ok {
			buckNum++
			bucketCounts[int(b["doc_count"].(float64))]++
		} else {
			err = fmt.Errorf("parsing json error")
			return err
		}
	}
	for _, bucket := range group["buckets"].([]interface{}) {
		if b, ok := bucket.(map[string]interface{}); ok {
			bucketCounts[int(b["doc_count"].(float64))]--
		} else {
			err = fmt.Errorf("parsing json error")
			return err
		}
	}
	for _, v := range bucketCounts {
		if v != 0 {
			err = fmt.Errorf("bucket size not match")
			return err
		}
	}
	fmt.Printf("\tquery return %v buckets\n", buckNum)
	return nil
}

func compareResp(respES *elastic.SearchResult, resp *elastic.SearchResult) error {
	if respES.Hits.TotalHits != resp.Hits.TotalHits {
		err := fmt.Errorf(`get %v documents, but %v expected`, resp.Hits.TotalHits, respES.Hits.TotalHits)
		return err
	}
	fmt.Printf("\tget %v documents, document number matches\n", respES.Hits.TotalHits)

	docIDES := make(map[string]int)
	if respES.TotalHits() > 0 {
		for _, hit := range respES.Hits.Hits {
			docIDES[hit.Id]++
		}
		// if orderby is specified, es sql dsl won't return document id
		if int64(len(docIDES)) < respES.TotalHits() {
			return nil
		}
		for _, hit := range resp.Hits.Hits {
			docIDES[hit.Id]--
		}
		for _, v := range docIDES {
			if v != 0 {
				err := fmt.Errorf(`document id not match`)
				return err
			}
		}
	}
	return nil
}

func filter1(colName string) bool {
	return strings.Contains(colName, "col")
}

func replace(colName string) (string, error) {
	return "myCol" + colName[3:], nil
}

func filter2(colName string) bool {
	return strings.Contains(colName, "col")
}

func process(v string) (string, error) {
	return v, nil
}

func TestCoverage(t *testing.T) {
	e := NewESql()
	sqls, err := readSQLs(testCases)
	if err != nil {
		t.Errorf("Fail to load testcases")
		return
	}

	for i, sql := range sqls {
		fmt.Printf("test %dth query ...\n", i+1)
		_, _, err := e.ConvertPretty(sql)
		if err != nil {
			t.Errorf("%vth query fails: %v", i+1, err)
			return
		}
	}

	sqls, err = readSQLs(testCasesInvalid)
	if err != nil {
		t.Errorf("Fail to load testcasesInvalid")
		return
	}

	for i, sql := range sqls {
		fmt.Printf("test %dth query ...\n", i+1)
		_, _, err := e.ConvertPretty(sql)
		if err == nil {
			t.Errorf("%vth query should fail but not", i+1)
			return
		}
	}

	e.SetDefault()
	e.SetCadence(true)
	e.SetReplace(filter1, replace)
	e.SetProcess(filter2, process)
	e.SetPageSize(1000)
	e.SetBucketNum(500)

	sqls, err = readSQLs(testCasesCad)
	if err != nil {
		t.Errorf("Fail to load testcasesCad")
		return
	}

	for i, sql := range sqls {
		fmt.Printf("test %dth query ...\n", i+1)
		_, _, err := e.ConvertPrettyCadence(sql, "1", "12")
		if err != nil {
			t.Errorf("%vth query fails: %v", i+1, err)
			return
		}
		fmt.Printf("%v\n", err)
	}

	sqls, err = readSQLs(testCasesInvalidCad)
	if err != nil {
		t.Errorf("Fail to load testcasesInvalidCad")
		return
	}

	for i, sql := range sqls {
		fmt.Printf("test %dth query ...\n", i+1)
		_, _, err := e.ConvertPrettyCadence(sql, "1", "12")
		if err == nil {
			t.Errorf("%vth query should fail but not", i+1)
			return
		}
		fmt.Printf("%v\n", err)
	}
}

func TestSQL(t *testing.T) {
	fmt.Println("Test SQLs ...")

	// initilizations
	var e ESql
	e.SetDefault()
	clientHTTP := &http.Client{Timeout: time.Second * 5}
	client, err := elastic.NewClient(elastic.SetURL(urlES))
	if err != nil {
		t.Error("Fail to create elastic client")
	}
	ctx := context.Background()

	sqls, err := readSQLs(testCases)
	if err != nil {
		t.Errorf("Fail to load testcases")
		return
	}

	f, err := os.Create(testDslsPretty)
	if err != nil {
		t.Error("Fail to create dslPretty file")
	}

	for i, sql := range sqls {
		fmt.Printf("testing %vth query...\n", i+1)
		// use es sql translate api to convert sql to dsl
		sqlQuery := fmt.Sprintf(`{"query": "%v"}`, sql)
		sqlReq, err := http.NewRequest("GET", urlSQL, bytes.NewBuffer([]byte(sqlQuery)))
		if err != nil {
			t.Errorf(`esql test: %vth sql translate query failed to create request struct: %v`, i+1, err)
		}
		sqlReq.Header.Add("Content-type", "application/json")
		sqlResp, err := clientHTTP.Do(sqlReq)
		if err != nil {
			t.Errorf(`esql test: %vth sql translate query failed: %v`, i+1, err)
		}
		sqlRespBody, err := ioutil.ReadAll(sqlResp.Body)
		if err != nil {
			t.Errorf(`esql test: %vth sql translate query failed to read body: %v`, i+1, err)
		}

		// use esql to translate sql to dsl
		dsl, _, err := e.ConvertPretty(sql)
		if err != nil {
			t.Errorf(`esql test: %vth query convert pretty fail: %v`, i+1, err)
		}
		f.WriteString("\n**************************\n" + strconv.Itoa(i+1) + "th query\n")
		f.WriteString(dsl)
		fmt.Printf("\tquery dsl generated\n")

		// query with esql dsl
		resp, err := client.Search(index).Source(dsl).Do(ctx)
		if err != nil {
			t.Errorf(`esql test: %vth ESQL DSL query fail: %v`, i+1, err)
		}
		if resp.Error != nil {
			t.Errorf(`esql test: %vth ESQL DSL query fail with error: %v`, i+1, resp.Error)
		}
		fmt.Printf("\tgenerated dsl is syntactically correct\n")

		skip := false
		for _, k := range notTestedKeywords {
			if strings.Contains(sql, k) {
				skip = true
				fmt.Printf("\tquery contains %v, not covered in test module\n", k)
				break
			}
		}
		if skip {
			continue
		}

		sqlDsl := string(sqlRespBody)
		respES, err := client.Search(index).Source(sqlDsl).Do(ctx)
		if err != nil || respES.Error != nil {
			fmt.Printf("\tresp get ERR %v, query not covered in test module\n", err)
			continue
		}

		// compare query results
		if strings.Contains(dsl, "aggs") {
			if strings.Contains(dsl, "groupby") {
				err = compareRespGroup(respES, resp)
			} else {
				fmt.Printf("\tquery contains aggregations without group by, not covered in test module\n")
				continue
			}
		} else {
			err = compareResp(respES, resp)
		}

		if err != nil {
			t.Errorf(`esql test: %vth query results not match, %v`, i+1, err)
		} else {
			fmt.Printf("\tpassed\n")
		}
	}
}

func myfilter(s string) bool {
	return s != "colE"
}

func TestGenCadenceDSL(t *testing.T) {
	fmt.Println("Test Generating DSL for Cadence...")
	var e ESql
	e.SetDefault()
	e.SetCadence(true)
	e.SetPageSize(100)
	e.SetBucketNum(100)
	//e.SetFilter(myfilter)
	//e.SetReplace(defaultCadenceColNameReplacePolicy)

	sqls, err := readSQLs(testCases)
	if err != nil {
		t.Errorf("Fail to load testcases")
		return
	}

	fp, err := os.Create(testDslsPrettyCadence)
	if err != nil {
		t.Error("Fail to create dsl file")
	}
	start := time.Now()
	for i, sql := range sqls {
		dslPretty, _, err := e.ConvertPretty(sql, "1", 123)
		if err != nil {
			t.Error(err)
		}
		fp.WriteString("\n**************************\n" + strconv.Itoa(i+1) + "th query\n")
		fp.WriteString(dslPretty)
		fmt.Printf("query %d dsl generated\n", i+1)
	}
	elapsed := time.Since(start)
	fmt.Printf("Time taken to generate all dsls: %s", elapsed)
	fp.Close()
	fmt.Println("DSL Cadence generated\n---------------------------------------------------------------------")
}

func readSQLs(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var sqls []string
	for scanner.Scan() {
		sqls = append(sqls, scanner.Text())
	}
	f.Close()
	return sqls, nil
}

// func testBenchmark(t *testing.T, choice string, round int) {
// 	var e ESql
// 	e.SetDefault()
// 	e.SetCadence(true)

// 	client, err := elastic.NewClient(elastic.SetURL(urlES))
// 	if err != nil {
// 		t.Error("Fail to create elastic client")
// 	}
// 	ctx := context.Background()

// 	var fileName string
// 	search := false
// 	switch choice {
// 	case "conversion":
// 		fileName = testCasesBenchmarkAll
// 	case "search":
// 		fileName = testCasesBenchmarkAll
// 		search = true
// 	default:
// 		t.Errorf("wrong choice")
// 		return
// 	}

// 	sqls, err := readSQLs(fileName)
// 	if err != nil {
// 		t.Errorf("Fail to open testcase file %v", fileName)
// 		return
// 	}
// 	var dslEsql, dslElasticsql []string
// 	var esqlResp, elasticsqlResp []*elastic.SearchResult
// 	start := time.Now()
// 	for i, sql := range sqls {
// 		for k := 0; k < round; k++ {
// 			dsl, _, err := e.Convert(sql)
// 			if err != nil {
// 				t.Errorf("Esql fail at %vth query: %v", i+1, err)
// 				return
// 			}
// 			if k == 0 {
// 				dslEsql = append(dslEsql, dsl)
// 			}
// 			if search {
// 				if resp, err := client.Search(index).Source(dsl).Do(ctx); err == nil {
// 					if k == 0 {
// 						esqlResp = append(esqlResp, resp)
// 					}
// 				} else {
// 					t.Errorf("Esql query fail at %vth query: %v", i+1, err)
// 					return
// 				}
// 			}
// 		}
// 	}
// 	elapsedEsql := time.Since(start)
// 	start = time.Now()
// 	for i, sql := range sqls {
// 		for k := 0; k < round; k++ {
// 			dsl, err := getCustomizedDSLFromSQL(sql, "0")
// 			if err != nil {
// 				t.Errorf("Elasticsql fail at %vth query", i+1)
// 				return
// 			}
// 			if k == 0 {
// 				dslElasticsql = append(dslElasticsql, dsl)
// 			}
// 			// fmt.Println(dsl)
// 			if search {
// 				if resp, err := client.Search(index).Source(dsl).Do(ctx); err == nil {
// 					if k == 0 {
// 						elasticsqlResp = append(elasticsqlResp, resp)
// 					}
// 				} else {
// 					t.Errorf("Elasticsql query fail at %vth query: %v", i+1, err)
// 					return
// 				}
// 			}
// 		}
// 	}

// 	elapsedElasticsql := time.Since(start)
// 	fmt.Printf("Convert %v sqls cost: esql: %v, elasticsql: %v\n", len(sqls)*round, elapsedEsql, elapsedElasticsql)
// 	f, _ := os.Create(testBenchmarkDslsEsql)
// 	for _, dsl := range dslEsql {
// 		f.WriteString(dsl + "\n")
// 	}
// 	f.Close()
// 	f, _ = os.Create(testBenchmarkDslsElasticsql)
// 	for _, dsl := range dslElasticsql {
// 		f.WriteString(dsl + "\n")
// 	}
// 	f.Close()
// 	// for i := range esqlResp {
// 	// 	if esqlResp[i].Hits.TotalHits != elasticsqlResp[i].Hits.TotalHits {
// 	// 		fmt.Printf("%vth query hits does not match, esql %v, elasticsql %v\n",
// 	// 			i+1, esqlResp[i].Hits.TotalHits, elasticsqlResp[i].Hits.TotalHits)
// 	// 	} else {
// 	// 		fmt.Printf("%vth query gets %v hits\n", i+1, elasticsqlResp[i].Hits.TotalHits)
// 	// 	}
// 	// }
// }

// func TestConversionSpeed(t *testing.T) {
// 	fmt.Println("Compare performance between esql and elasticsql  ... ")
// 	testBenchmark(t, "conversion", 1000)
// }

// func TestQueryProcessingSpeed(t *testing.T) {
// 	fmt.Println("Compare performance dsl processing in ES server between esql and elasticsql  ... ")
// 	testBenchmark(t, "search", 100)
// }
