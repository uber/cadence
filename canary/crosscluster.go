// Copyright (c) 2019 Uber Technologies, Inc.
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

package canary

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// some sample data to be passed around
type Data struct {
	Val        string
	Iterations int
}

func (c canaryImpl) crossClusterParentWf(ctx workflow.Context) error {
	// first try launching a child workflow in another cluster, active in another region
	logger := workflow.GetLogger(ctx)
	logger.Info("starting child workflow in domain1, cluster 1")
	ctx1 := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		Domain:                       c.getCrossClusterTargetDomain(),
		WorkflowID:                   "wf-domain-1-" + uuid.New().String(),
		TaskList:                     crossClusterDestTasklist,
		ExecutionStartToCloseTimeout: 1 * time.Minute,
	})
	err := workflow.ExecuteChildWorkflow(ctx1, wfTypeCrossClusterChild, Data{Val: "test"}).Get(ctx1, nil)
	if err != nil {
		logger.Error("got error executing child workflow", zap.Error(err))
		return err
	}
	logger.Info("test success - Cross-cluster cross-domain workflow completed", zap.Any("return-value", nil))

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    4 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	err = workflow.ExecuteActivity(ctx, c.failoverDestDomainActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("error during cross-cluster failover", zap.Error(err))
		return err
	}
	return nil
}

func (c canaryImpl) crossClusterChildWf(ctx workflow.Context, args Data) error {
	if args.Val != "test" {
		panic("wf1 did not receive expected args")
	}
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    4 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("workflow child wf starting activity.")
	err := workflow.ExecuteActivity(ctx, activityTypeCrossCluster).Get(ctx, nil)
	if err != nil {
		logger.Error("activity error", zap.Error(err))
	}
	// make the workflow do a loop
	if args.Iterations == 0 {
		logger.Info("continuing as new")
		return workflow.NewContinueAsNewError(ctx, c.crossClusterChildWf, Data{
			Val:        args.Val,
			Iterations: args.Iterations + 1,
		})
	}
	logger.Info("workflow child wf completed.")
	return nil
}

func (c canaryImpl) crossClusterSampleActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("activity 1 - running")
	return "Hello - activity 1", nil
}

// fails over the Dest domain so that both types of cross cluster operations can be tested
func (c canaryImpl) failoverDestDomainActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	srcReplicationInfo, err := getReplicationInfo(ctx, c.canaryClient, c.canaryDomain)
	fmt.Println("Src replication info", srcReplicationInfo)
	if err != nil {
		return err
	}
	destReplicationInfo, err := getReplicationInfo(ctx, *c.crossClusterDestClient, c.getCrossClusterTargetDomain())
	fmt.Println("dest replication info", srcReplicationInfo)
	if err != nil {
		return err
	}
	srcActiveCluster := srcReplicationInfo.GetActiveClusterName()
	destActiveCluster := destReplicationInfo.GetActiveClusterName()

	currentlyInSameCluster := srcActiveCluster == destActiveCluster
	// don't do this every iteration so as to not get rate-limited
	smallProbability := time.Now().Second() <= 10
	targetDomain := c.getCrossClusterTargetDomain()

	if currentlyInSameCluster && smallProbability {
		// do a failover to move the dest to a different cluster
		// thereby making the next iteration of the canary be cross-cluster AND cross domain
		// this operation is probably nondeterministic, but that's ok in this instance
		for _, cluster := range destReplicationInfo.GetClusters() {
			if cluster.GetClusterName() != srcActiveCluster {
				err = c.crossClusterDestClient.DomainClient.Update(ctx,
					&shared.UpdateDomainRequest{
						Name:                     &targetDomain,
						ReplicationConfiguration: &shared.DomainReplicationConfiguration{ActiveClusterName: cluster.ClusterName},
					})
				if err != nil {
					return fmt.Errorf("failed to update destination domain for failover in crossdomain canary test: cluster trying to move dest domain to %s, error: %w ", cluster.GetClusterName(), err)
				}
				logger.Info("failed over cross-cluster domain to ", zap.String("cluster", cluster.GetClusterName()))
				return nil
			}
		}
	} else if smallProbability {
		// else they're not currently in the same cluster, so fail the destination domain back to the same cluster
		// so the next iteration of the cross cluster canary test will be just cross domain
		err = c.crossClusterDestClient.DomainClient.Update(ctx,
			&shared.UpdateDomainRequest{
				ReplicationConfiguration: &shared.DomainReplicationConfiguration{ActiveClusterName: &srcActiveCluster},
				Name:                     &targetDomain,
			})
		if err != nil {
			return fmt.Errorf("failed to update destination domain for failover in crossdomain canary test - in this instance the dest back to the same cluster as the source: %w", err)
		}
		logger.Info("failed over cross-cluster domain to ", zap.String("cluster", srcActiveCluster))
	}
	return nil
}

func getReplicationInfo(ctx context.Context, client cadenceClient, domain string) (*shared.DomainReplicationConfiguration, error) {
	res, err := client.DomainClient.Describe(ctx, domain)
	if err != nil {
		return nil, err
	}
	return res.GetReplicationConfiguration(), nil
}
