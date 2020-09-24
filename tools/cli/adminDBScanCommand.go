// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/reconciliation/fetcher"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/worker/scanner/executions"
)

// AdminDBScan is used to scan over executions in database and detect corruptions.
func AdminDBScan(c *cli.Context) {
	scanType, err := executions.ScanTypeString(c.String(FlagScanType))

	if err != nil {
		ErrorAndExit("unknown scan type", err)
	}

	numberOfShards := c.Int(FlagNumberOfShards)
	collectionSlice := c.StringSlice(FlagInvariantCollection)

	var collections []invariant.Collection
	for _, v := range collectionSlice {
		collection, err := invariant.CollectionString(v)
		if err != nil {
			ErrorAndExit("unknown invariant collection", err)
		}
		collections = append(collections, collection)
	}

	invariants := scanType.ToInvariants(collections)
	if len(invariants) < 1 {
		ErrorAndExit(
			fmt.Sprintf("no invariants for scan type %q and collections %q",
				scanType.String(),
				collectionSlice),
			nil,
		)
	}
	ef := scanType.ToExecutionFetcher()

	input := getInputFile(c.String(FlagInputFile))

	dec := json.NewDecoder(input)
	if err != nil {
		ErrorAndExit("", err)
	}
	var data []fetcher.ExecutionRequest

	for {
		var exec fetcher.ExecutionRequest
		if err := dec.Decode(&exec); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
		} else {
			data = append(data, exec)
		}
	}

	for _, e := range data {
		execution, result := checkExecution(c, numberOfShards, e, invariants, ef)
		out := store.ScanOutputEntity{
			Execution: execution,
			Result:    result,
		}
		data, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		fmt.Println(string(data))
	}
}

func checkExecution(
	c *cli.Context,
	numberOfShards int,
	req fetcher.ExecutionRequest,
	invariants []executions.InvariantFactory,
	fetcher executions.ExecutionFetcher,
) (interface{}, invariant.ManagerCheckResult) {
	session := connectToCassandra(c)
	defer session.Close()
	logger := loggerimpl.NewNopLogger()

	execStore, err := cassandra.NewWorkflowExecutionPersistence(
		common.WorkflowIDToHistoryShard(req.WorkflowID, numberOfShards),
		session,
		logger,
	)

	if err != nil {
		ErrorAndExit("Failed to get execution store", err)
	}

	historyV2Mgr := persistence.NewHistoryV2ManagerImpl(
		cassandra.NewHistoryV2PersistenceFromSession(session, logger),
		logger,
		dynamicconfig.GetIntPropertyFn(common.DefaultTransactionSizeLimit),
	)

	pr := persistence.NewPersistenceRetryer(
		persistence.NewExecutionManagerImpl(execStore, logger),
		historyV2Mgr,
		common.CreatePersistenceRetryPolicy(),
	)

	execution, err := fetcher(pr, req)

	if err != nil {
		return nil, invariant.ManagerCheckResult{}
	}

	var ivs []invariant.Invariant

	for _, fn := range invariants {
		ivs = append(ivs, fn(pr))
	}

	return execution, invariant.NewInvariantManager(ivs).RunChecks(execution)
}
