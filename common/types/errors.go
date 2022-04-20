// Copyright (c) 2020 Uber Technologies Inc.
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

package types

import "github.com/uber/cadence/common/log/tag"

func (err AccessDeniedError) Error() string {
	return err.Message
}

func (err BadRequestError) Error() string {
	return err.Message
}

func (err CancellationAlreadyRequestedError) Error() string {
	return err.Message
}

func (err ClientVersionNotSupportedError) Error() string {
	return "client version not supported"
}

func (err ClientVersionNotSupportedError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.FeatureVersion(err.FeatureVersion),
		tag.ClientImplementation(err.ClientImpl),
		tag.SupportedVersions(err.SupportedVersions),
	}
}

func (err FeatureNotEnabledError) Error() string {
	return "feature not enabled"
}

func (err FeatureNotEnabledError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.FeatureFlag(err.FeatureFlag),
	}
}

func (err CurrentBranchChangedError) Error() string {
	return err.Message
}

func (err CurrentBranchChangedError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.WorkflowBranchToken(string(err.CurrentBranchToken)),
	}
}

func (err DomainAlreadyExistsError) Error() string {
	return err.Message
}

func (err DomainNotActiveError) Error() string {
	return err.Message
}

func (err DomainNotActiveError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.WorkflowDomainName(err.DomainName),
		tag.ClusterName(err.CurrentCluster),
		tag.ActiveClusterName(err.ActiveCluster),
	}
}

func (err EntityNotExistsError) Error() string {
	return err.Message
}

func (err EntityNotExistsError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.ClusterName(err.CurrentCluster),
		tag.ActiveClusterName(err.ActiveCluster),
	}
}

func (err WorkflowExecutionAlreadyCompletedError) Error() string {
	return err.Message
}

func (err InternalDataInconsistencyError) Error() string {
	return err.Message
}

func (err InternalServiceError) Error() string {
	return err.Message
}

func (err LimitExceededError) Error() string {
	return err.Message
}

func (err QueryFailedError) Error() string {
	return err.Message
}

func (err RemoteSyncMatchedError) Error() string {
	return err.Message
}

func (err RetryTaskV2Error) Error() string {
	return err.Message
}

func (err RetryTaskV2Error) Tags() []tag.Tag {
	tags := []tag.Tag{
		tag.WorkflowDomainID(err.DomainID),
		tag.WorkflowID(err.WorkflowID),
		tag.WorkflowRunID(err.RunID),
	}
	if err.StartEventID != nil {
		tags = append(tags, tag.WorkflowStartEventID(*err.StartEventID))
	}
	if err.StartEventVersion != nil {
		tags = append(tags, tag.WorkflowStartEventVersion(*err.StartEventVersion))
	}
	if err.EndEventID != nil {
		tags = append(tags, tag.WorkflowEndEventID(*err.EndEventID))
	}
	if err.EndEventVersion != nil {
		tags = append(tags, tag.WorkflowEndEventVersion(*err.EndEventVersion))
	}
	return tags
}

func (err ServiceBusyError) Error() string {
	return err.Message
}

func (err WorkflowExecutionAlreadyStartedError) Error() string {
	return err.Message
}

func (err WorkflowExecutionAlreadyStartedError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.RequestID(err.StartRequestID),
		tag.WorkflowRunID(err.RunID),
	}
}

func (err ShardOwnershipLostError) Error() string {
	return err.Message
}

func (err ShardOwnershipLostError) Tags() []tag.Tag {
	return []tag.Tag{
		tag.ShardOwner(err.Owner),
	}
}

func (err EventAlreadyStartedError) Error() string {
	return err.Message
}
