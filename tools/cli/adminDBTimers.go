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
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/quotas"
)

// AdminTimers is used to list scheduled timers.
func AdminTimers(c *cli.Context) {
	lowerShardBound := c.Int(FlagLowerShardBound)
	upperShardBound := c.Int(FlagUpperShardBound)
	startDate := c.String(FlagStartDate)
	endDate := c.String(FlagEndDate)
	rps := c.Int(FlagRPS)

	ctx, cancel := newContextForLongPoll(c)
	defer cancel()

	st, err := parseSingleTs(startDate)
	if err != nil {
		ErrorAndExit("wrong date format for "+FlagStartDate, err)
	}
	et, err := parseSingleTs(endDate)
	if err != nil {
		ErrorAndExit("wrong date format for "+FlagEndDate, err)
	}

	session := connectToCassandra(c)
	defer session.Close()

	// shared limiter
	limiter := quotas.NewSimpleRateLimiter(rps)

	group, ctx := errgroup.WithContext(ctx)
	for shardID := lowerShardBound; shardID < upperShardBound; shardID++ {
		shardID := shardID
		group.Go(func() error {
			return printTimers(c, session, shardID, st, et, limiter)
		})
	}

	if err := group.Wait(); err != nil {
		ErrorAndExit("getting timers", err)
	}
}

func printTimers(
	c *cli.Context,
	session *gocql.Session,
	shardID int,
	startDate, endDate time.Time,
	limiter *quotas.RateLimiter,
) error {
	domainID := c.String(FlagDomainID)
	skipErrMode := c.Bool(FlagSkipErrorMode)
	batchSize := c.Int(FlagBatchSize)
	timerTypes := intSliceToMap(c.IntSlice(FlagTimerType))
	logger := loggerimpl.NewNopLogger()

	execStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, logger)
	if err != nil {
		err = errors.Wrapf(err, "cannot get execution manager for shardId %d", shardID)
		if !skipErrMode {
			return err
		}
		fmt.Println(err.Error())
		return nil
	}

	manager := persistence.NewExecutionManagerImpl(execStore, logger)
	ratelimitedClient := persistence.NewWorkflowExecutionPersistenceRateLimitedClient(manager, limiter, logger)

	var token []byte
	isFirstIteration := true
	for isFirstIteration || len(token) != 0 {
		isFirstIteration = false
		req := persistence.GetTimerIndexTasksRequest{
			MinTimestamp:  startDate,
			MaxTimestamp:  endDate,
			BatchSize:     batchSize,
			NextPageToken: token,
		}

		resp := &persistence.GetTimerIndexTasksResponse{}

		op := func() error {
			var err error
			resp, err = ratelimitedClient.GetTimerIndexTasks(&req)
			return err
		}

		for attempt := 0; attempt < maxDBRetries; attempt++ {
			err = backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
			if err == nil {
				break
			}
		}

		if err != nil {
			err := errors.Wrapf(err, "cannot get timer tasks for shardId %d", shardID)
			if !skipErrMode {
				return err
			}

			fmt.Println(err.Error())
			return nil
		}

		token = resp.NextPageToken

		for _, t := range resp.Timers {
			if len(domainID) > 0 && t.DomainID != domainID {
				continue
			}

			if _, ok := timerTypes[t.TaskType]; !ok {
				continue
			}

			data, err := json.Marshal(t)
			if err != nil {
				err = errors.Wrap(err, "cannot marshal timer to json")
				if !skipErrMode {
					return err
				}
				fmt.Println(err.Error())
			} else {
				fmt.Println(string(data))
			}
		}
	}

	return nil
}
