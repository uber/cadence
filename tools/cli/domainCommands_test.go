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
	"flag"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/types"
)

type domainCommandsSuite struct {
	suite.Suite
	mockCtrl             *gomock.Controller
	serverFrontendClient *frontend.MockClient
	domainHandler        *domain.MockHandler
	cliCtx               *cli.Context
	domainCLI            *domainCLIImpl
}

func TestDomainCommandsSuite(t *testing.T) {
	s := new(domainCommandsSuite)
	suite.Run(t, s)
}

func (s *domainCommandsSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())

	s.serverFrontendClient = frontend.NewMockClient(s.mockCtrl)
	s.domainHandler = domain.NewMockHandler(s.mockCtrl)

	set := flag.NewFlagSet("test", 0)
	set.Int(FlagContextTimeout, 10, "test flag")
	s.cliCtx = cli.NewContext(nil, set, nil)
	s.domainCLI = &domainCLIImpl{
		s.serverFrontendClient,
		s.domainHandler,
	}
}

func (s *domainCommandsSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *domainCommandsSuite) TestRebalanceDomains_Success() {
	listDomainsResp := getListDomainsResponse()
	preferredCluster := "active"
	s.serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResp, nil)
	expectedUpdateRequest := &types.UpdateDomainRequest{
		Name:              listDomainsResp.Domains[0].GetDomainInfo().GetName(),
		ActiveClusterName: &preferredCluster,
	}
	s.serverFrontendClient.EXPECT().
		UpdateDomain(gomock.Any(), expectedUpdateRequest).
		Return(&types.UpdateDomainResponse{}, nil).Times(1)
	success, fail := s.domainCLI.rebalanceDomains(s.cliCtx)
	s.Equal(1, len(success))
	s.Equal(0, len(fail))
}

func (s *domainCommandsSuite) TestRebalanceDomains_UpdateDomain_Error() {
	listDomainsResp := getListDomainsResponse()
	preferredCluster := "active"
	s.serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResp, nil)
	expectedUpdateRequest := &types.UpdateDomainRequest{
		Name:              listDomainsResp.Domains[0].GetDomainInfo().GetName(),
		ActiveClusterName: &preferredCluster,
	}
	s.serverFrontendClient.EXPECT().
		UpdateDomain(gomock.Any(), expectedUpdateRequest).
		Return(nil, fmt.Errorf("test")).Times(1)
	success, fail := s.domainCLI.rebalanceDomains(s.cliCtx)
	s.Equal(0, len(success))
	s.Equal(1, len(fail))
}

func (s *domainCommandsSuite) TestRebalanceDomains_IsDryRun_Success() {
	set := flag.NewFlagSet("test", 0)
	set.Int(FlagContextTimeout, 10, "test flag")
	set.Bool(FlagDryRun, true, "dry run")
	cliCtx := cli.NewContext(nil, set, nil)
	listDomainsResp := getListDomainsResponse()
	s.serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResp, nil)
	s.serverFrontendClient.EXPECT().
		UpdateDomain(gomock.Any(), gomock.Any()).
		Return(&types.UpdateDomainResponse{}, nil).Times(0)
	success, fail := s.domainCLI.rebalanceDomains(cliCtx)
	s.Equal(1, len(success))
	s.Equal(0, len(fail))
}

func (s *domainCommandsSuite) TestRebalanceDomains_NoManagedFailover_Skipped() {
	listDomainsResp := getListDomainsResponse()
	preferredCluster := "active"
	listDomainsResp.Domains[0].DomainInfo.Data = map[string]string{
		common.DomainDataKeyForManagedFailover:  "false",
		common.DomainDataKeyForPreferredCluster: preferredCluster,
	}
	s.serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResp, nil)

	s.serverFrontendClient.EXPECT().
		UpdateDomain(gomock.Any(), gomock.Any()).
		Return(&types.UpdateDomainResponse{}, nil).Times(0)
	success, fail := s.domainCLI.rebalanceDomains(s.cliCtx)
	s.Equal(0, len(success))
	s.Equal(0, len(fail))
}

func (s *domainCommandsSuite) TestRebalanceDomains_NoPreferredCluster_Skipped() {
	listDomainsResp := getListDomainsResponse()
	listDomainsResp.Domains[0].DomainInfo.Data = map[string]string{
		common.DomainDataKeyForManagedFailover: "true",
	}
	s.serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResp, nil)
	s.serverFrontendClient.EXPECT().
		UpdateDomain(gomock.Any(), gomock.Any()).
		Return(&types.UpdateDomainResponse{}, nil).Times(0)
	success, fail := s.domainCLI.rebalanceDomains(s.cliCtx)
	s.Equal(0, len(success))
	s.Equal(0, len(fail))
}

func getListDomainsResponse() *types.ListDomainsResponse {
	return &types.ListDomainsResponse{
		Domains: []*types.DescribeDomainResponse{
			{
				DomainInfo: &types.DomainInfo{
					Name:   uuid.NewUUID().String(),
					Status: types.DomainStatusRegistered.Ptr(),
					Data: map[string]string{
						common.DomainDataKeyForManagedFailover:  "true",
						common.DomainDataKeyForPreferredCluster: "active",
					},
				},
				IsGlobalDomain: true,
			},
		},
		NextPageToken: nil,
	}
}
