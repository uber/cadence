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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/persistence/serialization"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/reconciliation/fetcher"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/worker/scanner/executions"
)

const (
	listContextTimeout = time.Minute
)

// AdminDBScan is used to scan over executions in database and detect corruptions.
func AdminDBScan(c *cli.Context) {
	scanType, err := executions.ScanTypeString(c.String(FlagScanType))

	if err != nil {
		ErrorAndExit("unknown scan type", err)
	}

	numberOfShards := getRequiredIntOption(c, FlagNumberOfShards)
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
	client, session := connectToCassandra(c)
	defer session.Close()
	logger := loggerimpl.NewNopLogger()

	execStore, err := cassandra.NewWorkflowExecutionPersistence(
		common.WorkflowIDToHistoryShard(req.WorkflowID, numberOfShards),
		client,
		session,
		logger,
	)

	if err != nil {
		ErrorAndExit("Failed to get execution store", err)
	}

	historyV2Mgr := persistence.NewHistoryV2ManagerImpl(
		cassandra.NewHistoryV2PersistenceFromSession(client, session, logger),
		logger,
		dynamicconfig.GetIntPropertyFn(common.DefaultTransactionSizeLimit),
	)

	pr := persistence.NewPersistenceRetryer(
		persistence.NewExecutionManagerImpl(execStore, logger),
		historyV2Mgr,
		common.CreatePersistenceRetryPolicy(),
	)

	ctx, cancel := newContext(c)
	defer cancel()

	execution, err := fetcher(ctx, pr, req)

	if err != nil {
		return nil, invariant.ManagerCheckResult{}
	}

	var ivs []invariant.Invariant

	for _, fn := range invariants {
		ivs = append(ivs, fn(pr))
	}

	return execution, invariant.NewInvariantManager(ivs).RunChecks(ctx, execution)
}

// AdminDBScanUnsupportedWorkflow is to scan DB for unsupported workflow for a new release
func AdminDBScanUnsupportedWorkflow(c *cli.Context) {
	rps := c.Int(FlagRPS)
	outputFile := getOutputFile(c.String(FlagOutputFilename))
	startShardID := c.Int(FlagLowerShardBound)
	endShardID := c.Int(FlagUpperShardBound)

	defer outputFile.Close()
	for i := startShardID; i <= endShardID; i++ {
		listExecutionsByShardID(c, i, rps, outputFile)
		fmt.Println(fmt.Sprintf("Shard %v scan operation is completed.", i))
	}
}

func listExecutionsByShardID(
	c *cli.Context,
	shardID int,
	rps int,
	outputFile *os.File,
) {

	client := initializeExecutionStore(c, shardID, rps)
	defer client.Close()

	paginationFunc := func(paginationToken []byte) ([]interface{}, []byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), listContextTimeout)
		defer cancel()

		resp, err := client.ListConcreteExecutions(
			ctx,
			&persistence.ListConcreteExecutionsRequest{
				PageSize:  1000,
				PageToken: paginationToken,
			},
		)
		if err != nil {
			return nil, nil, err
		}
		var paginateItems []interface{}
		for _, history := range resp.Executions {
			paginateItems = append(paginateItems, history)
		}
		return paginateItems, resp.PageToken, nil
	}

	executionIterator := collection.NewPagingIterator(paginationFunc)
	for executionIterator.HasNext() {
		result, err := executionIterator.Next()
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Failed to scan shard ID: %v for unsupported workflow. Please retry.", shardID), err)
		}
		execution := result.(*persistence.ListConcreteExecutionsEntity)
		executionInfo := execution.ExecutionInfo
		if executionInfo != nil && executionInfo.CloseStatus == 0 && execution.VersionHistories == nil {

			outStr := fmt.Sprintf("cadence --address <host>:<port> --domain <%v> workflow reset --wid %v --rid %v --reset_type LastDecisionCompleted --reason 'release 0.16 upgrade'\n",
				executionInfo.DomainID,
				executionInfo.WorkflowID,
				executionInfo.RunID,
			)
			if _, err = outputFile.WriteString(outStr); err != nil {
				ErrorAndExit("Failed to write data to file", err)
			}
			if err = outputFile.Sync(); err != nil {
				ErrorAndExit("Failed to sync data to file", err)
			}
		}
	}
}

func initializeExecutionStore(
	c *cli.Context,
	shardID int,
	rps int,
) persistence.ExecutionManager {

	var execStore persistence.ExecutionStore
	dbType := c.String(FlagDBType)
	logger := loggerimpl.NewNopLogger()
	switch dbType {
	case "cassandra":
		execStore = initializeCassandraExecutionClient(c, shardID, logger)
	case "mysql":
		execStore = initializeSQLExecutionStore(c, shardID, logger)
	case "postgres":
		execStore = initializeSQLExecutionStore(c, shardID, logger)
	default:
		ErrorAndExit("The DB type is not supported. Options are: cassandra, mysql, postgres.", nil)
	}

	historyManager := persistence.NewExecutionManagerImpl(execStore, logger)
	rateLimiter := quotas.NewSimpleRateLimiter(rps)
	return persistence.NewWorkflowExecutionPersistenceRateLimitedClient(historyManager, rateLimiter, logger)
}

func initializeCassandraExecutionClient(
	c *cli.Context,
	shardID int,
	logger log.Logger,
) persistence.ExecutionStore {

	client, session := connectToCassandra(c)
	execStore, err := cassandra.NewWorkflowExecutionPersistence(
		shardID,
		client,
		session,
		logger,
	)
	if err != nil {
		ErrorAndExit("Failed to get execution store from cassandra config", err)
	}
	return execStore
}

func initializeSQLExecutionStore(
	c *cli.Context,
	shardID int,
	logger log.Logger,
) persistence.ExecutionStore {

	sqlDB := connectToSQL(c)
	encodingType := c.String(FlagEncodingType)
	decodingTypesStr := c.StringSlice(FlagDecodingTypes)
	var decodingTypes []common.EncodingType
	for _, dt := range decodingTypesStr {
		decodingTypes = append(decodingTypes, common.EncodingType(dt))
	}
	execStore, err := sql.NewSQLExecutionStore(sqlDB, logger, shardID, getSQLParser(common.EncodingType(encodingType), decodingTypes...))
	if err != nil {
		ErrorAndExit("Failed to get execution store from cassandra config", err)
	}
	return execStore
}

func getSQLParser(encodingType common.EncodingType, decodingTypes ...common.EncodingType) serialization.Parser {
	parser, err := serialization.NewParser(encodingType, decodingTypes...)
	if err != nil {
		ErrorAndExit("failed to initialize sql parser", err)
	}
	return parser
}
