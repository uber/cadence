// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package failoverManager

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
)

const (
	failoverManagerContextKey = "failoverManagerContext"
	// TaskListName tasklist
	TaskListName = "cadence-sys-failoverManager-tasklist"
	// WorkflowTypeName workflow type name
	WorkflowTypeName = "cadence-sys-failoverManager-workflow"
	// WorkflowID will be reused to ensure only one workflow running
	WorkflowID             = "cadence-failover-manager"
	failoverActivityName   = "cadence-sys-failover-activity"
	getDomainsActivityName = "cadence-sys-getDomains-activity"

	defaultBatchFailoverSize              = 10
	defaultBatchFailoverWaitTimeInSeconds = 10

	errMsgParamsIsNil          = "params is nil"
	errMsgTargetClusterIsEmpty = "targetCluster is empty"

	// QueryType for failover workflow
	QueryType    = "state"
	// PauseSignal signal name for pause
	PauseSignal  = "pause"
	// ResumeSignal signal name for resume
	ResumeSignal = "resume"
)

type (
	// FailoverParams is the arg for failoverWorkflow
	FailoverParams struct {
		// TargetCluster is the destination of failover
		TargetCluster string
		// BatchFailoverSize is number of domains to failover in one batch
		BatchFailoverSize int
		// BatchFailoverWaitTimeInSeconds is the waiting time between batch failover
		BatchFailoverWaitTimeInSeconds int
	}

	// FailoverResult is workflow result
	FailoverResult struct {
		SuccessDomains []string
		FailedDomains  []string
	}

	// GetDomainsActivityParams params for activity
	GetDomainsActivityParams struct {
		TargetCluster string
	}

	// FailoverActivityParams params for activity
	FailoverActivityParams struct {
		Domains       []string
		TargetCluster string
	}

	// FailoverActivityResult result for failover activity
	FailoverActivityResult struct {
		SuccessDomains []string
		FailedDomains  []string
	}

	// QueryResult ..
	QueryResult struct {
		TotalDomains  int
		Success       int
		Failed        int
		FailedDomains []string
	}
)

func init() {
	workflow.RegisterWithOptions(FailoverWorkflow, workflow.RegisterOptions{Name: WorkflowTypeName})
	activity.RegisterWithOptions(FailoverActivity, activity.RegisterOptions{Name: failoverActivityName})
	activity.RegisterWithOptions(GetDomainsActivity, activity.RegisterOptions{Name: getDomainsActivityName})
}

// FailoverWorkflow is the workflow that managed failover all domains with IsManagedByCadence=true
func FailoverWorkflow(ctx workflow.Context, params *FailoverParams) (*FailoverResult, error) {
	err := validateParams(params)
	if err != nil {
		return nil, err
	}

	ao := workflow.WithActivityOptions(ctx, getGetDomainsActivityOptions())
	getDomainsParams := &GetDomainsActivityParams{
		TargetCluster: params.TargetCluster,
	}
	var domains []string
	err = workflow.ExecuteActivity(ao, GetDomainsActivity, getDomainsParams).Get(ctx, &domains)
	if err != nil {
		return nil, err
	}

	var failedDomains []string
	var successDomains []string
	totalNumOfDomains := len(domains)
	err = workflow.SetQueryHandler(ctx, QueryType, func(input []byte) (*QueryResult, error) {
		return &QueryResult{
			TotalDomains:  totalNumOfDomains,
			Success:       len(successDomains),
			Failed:        len(failedDomains),
			FailedDomains: failedDomains,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	ao = workflow.WithActivityOptions(ctx, getFailoverActivityOptions())
	batchSize := params.BatchFailoverSize
	times := len(domains) / batchSize

	pauseCh := workflow.GetSignalChannel(ctx, PauseSignal)
	resumeCh := workflow.GetSignalChannel(ctx, ResumeSignal)
	var shouldPause bool

	for i := 0; i < times; i++ {
		// check if need to pause
		shouldPause = pauseCh.ReceiveAsync(nil)
		if shouldPause {
			// todo: add log workflow paused, and query state
			resumeCh.Receive(ctx, nil)
		}

		// failover domains
		failoverActivityParams := &FailoverActivityParams{
			Domains:       domains[i*batchSize : min((i+1)*batchSize, totalNumOfDomains)],
			TargetCluster: params.TargetCluster,
		}
		var actResult FailoverActivityResult
		workflow.ExecuteActivity(ao, FailoverActivity, failoverActivityParams).Get(ctx, &actResult)
		failedDomains = append(failedDomains, actResult.FailedDomains...)
		workflow.Sleep(ctx, time.Duration(params.BatchFailoverWaitTimeInSeconds)*time.Second)
	}

	return &FailoverResult{
		SuccessDomains: successDomains,
		FailedDomains:  failedDomains,
	}, nil
}

func getGetDomainsActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Second,
		StartToCloseTimeout:    20 * time.Second,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          2 * time.Second,
			BackoffCoefficient:       2,
			MaximumInterval:          1 * time.Minute,
			ExpirationInterval:       10 * time.Minute,
			NonRetriableErrorReasons: []string{errMsgParamsIsNil, errMsgTargetClusterIsEmpty},
		},
	}
}

func getFailoverActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Second,
		StartToCloseTimeout:    10 * time.Second,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func validateParams(params *FailoverParams) error {
	if params == nil {
		return errors.New(errMsgParamsIsNil)
	}
	if len(params.TargetCluster) == 0 {
		return errors.New(errMsgTargetClusterIsEmpty)
	}
	if params.BatchFailoverSize <= 0 {
		params.BatchFailoverSize = defaultBatchFailoverSize
	}
	if params.BatchFailoverWaitTimeInSeconds <= 0 {
		params.BatchFailoverWaitTimeInSeconds = defaultBatchFailoverWaitTimeInSeconds
	}
	return nil
}

// GetDomainsActivity activity def
func GetDomainsActivity(ctx context.Context, params *GetDomainsActivityParams) ([]string, error) {
	err := validateGetDomainsActivityParams(params)
	if err != nil {
		return nil, err
	}
	domains, err := getAllDomains(ctx)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, domain := range domains {
		if shouldFailover(domain, params.TargetCluster) {
			domainName := domain.GetDomainInfo().GetName()
			res = append(res, domainName)
		}
	}
	return res, nil
}

func validateGetDomainsActivityParams(params *GetDomainsActivityParams) error {
	if params == nil {
		return errors.New(errMsgParamsIsNil)
	}
	if len(params.TargetCluster) == 0 {
		return errors.New(errMsgTargetClusterIsEmpty)
	}
	return nil
}

func shouldFailover(domain *shared.DescribeDomainResponse, targetCluster string) bool {
	isDomainNotActiveInTargetCluster := domain.ReplicationConfiguration.GetActiveClusterName() != targetCluster
	return isDomainNotActiveInTargetCluster && isDomainFailoverManagedByCadence(domain)

}

func isDomainFailoverManagedByCadence(domain *shared.DescribeDomainResponse) bool {
	domainData := domain.DomainInfo.GetData()
	return strings.ToLower(strings.TrimSpace(domainData[common.DomainDataKeyForManagedFailover])) == "true"
}

func getClient(ctx context.Context) frontend.Client {
	manager := ctx.Value(failoverManagerContextKey).(*FailoverManager)
	feClient := manager.clientBean.GetFrontendClient()
	return feClient
}

func getAllDomains(ctx context.Context) ([]*shared.DescribeDomainResponse, error) {
	feClient := getClient(ctx)
	var res []*shared.DescribeDomainResponse
	pagesize := int32(200)
	var token []byte
	for more := true; more; more = len(token) > 0 {
		listRequest := &shared.ListDomainsRequest{
			PageSize:      common.Int32Ptr(pagesize),
			NextPageToken: token,
		}
		listResp, err := feClient.ListDomains(ctx, listRequest)
		if err != nil {
			return nil, err
		}
		token = listResp.GetNextPageToken()
		res = append(res, listResp.GetDomains()...)
		activity.RecordHeartbeat(ctx, len(res))
	}
	return res, nil
}

// FailoverActivity activity def
func FailoverActivity(ctx context.Context, params *FailoverActivityParams) (*FailoverActivityResult, error) {
	feClient := getClient(ctx)
	domains := params.Domains
	var successDomains []string
	var failedDomains []string
	for _, domain := range domains {
		replicationConfig := &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(params.TargetCluster),
		}
		updateRequest := &shared.UpdateDomainRequest{
			Name:                     common.StringPtr(domain),
			ReplicationConfiguration: replicationConfig,
		}
		_, err := feClient.UpdateDomain(ctx, updateRequest)
		if err != nil {
			failedDomains = append(failedDomains, domain)
		} else {
			successDomains = append(successDomains, domain)
		}
	}
	return &FailoverActivityResult{
		SuccessDomains: successDomains,
		FailedDomains:  failedDomains,
	}, nil
}
