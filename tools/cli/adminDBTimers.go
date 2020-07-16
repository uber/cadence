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

	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
)

// AdminTimers is used to list scheduled timers.
func AdminTimers(c *cli.Context) {
	lowerShardBound := c.Int(FlagLowerShardBound)
	upperShardBound := c.Int(FlagUpperShardBound)
	concurrency := c.Int(FlagConcurrency)
	batchSize := c.Int(FlagBatchSize)
	startDate := c.String(FlagStartDate)
	endDate := c.String(FlagEndDate)
	domainId := c.String(FlagDomainID)
	skipErrMode := c.Bool(FlagSkipErrorMode)

	session := connectToCassandra(c)
	defer session.Close()

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

	group, ctx := errgroup.WithContext(ctx)
	throttle := make(chan struct{}, concurrency)
	for shardID := lowerShardBound; shardID < upperShardBound; shardID++ {
		shardID := shardID
		group.Go(func() error {
			throttle <- struct{}{}
			defer func() {
				<-throttle
			}()

			return printTimers(session, shardID, batchSize, st, et, domainId, skipErrMode)
		})
	}

	if err := group.Wait(); err != nil {
		ErrorAndExit("getting timers", err)
	}
}

func printTimers(
	session *gocql.Session,
	shardID, batchSize int,
	startDate, endDate time.Time,
	domainId string,
	skipErrMode bool,
) error {
	execStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, loggerimpl.NewNopLogger())

	if err != nil {
		err = errors.Wrapf(err, "cannot get execution manager for shardId %d", shardID)
		if !skipErrMode {
			return err
		}
		fmt.Println(err.Error())
		return nil
	}

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
		resp, err := execStore.GetTimerIndexTasks(&req)

		if err != nil {
			return errors.Wrapf(err, "cannot get timer tasks for shardId %d", shardID)
		}

		token = resp.NextPageToken

		for _, t := range resp.Timers {
			if len(domainId) > 0 && t.DomainID != domainId {
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
