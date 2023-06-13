// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package isolationgroupapi

import (
	"context"

	"github.com/uber/cadence/common/types"
)

func (z *handlerImpl) GetDomainState(ctx context.Context, request types.GetDomainIsolationGroupsRequest) (*types.GetDomainIsolationGroupsResponse, error) {
	res, err := z.domainHandler.DescribeDomain(ctx, &types.DescribeDomainRequest{
		Name: &request.Domain,
	})
	if err != nil {
		return nil, err
	}
	if res == nil || res.Configuration == nil || res.Configuration.IsolationGroups == nil {
		return &types.GetDomainIsolationGroupsResponse{}, nil
	}
	return &types.GetDomainIsolationGroupsResponse{
		IsolationGroups: *res.Configuration.IsolationGroups,
	}, nil
}

// UpdateDomainState is the read operation for updating a domain's isolation-groups
// todo (david.porter) delete this handler and use domain-handler directly
func (z *handlerImpl) UpdateDomainState(ctx context.Context, request types.UpdateDomainIsolationGroupsRequest) error {
	err := z.domainHandler.UpdateIsolationGroups(ctx, request)
	return err
}
