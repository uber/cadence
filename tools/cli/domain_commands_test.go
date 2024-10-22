package cli

import (
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func (s *cliAppSuite) TestDomainRegister() {
	s.testcaseHelper([]testcase{
		{
			"local",
			"cadence --do test-domain domain register --global_domain false",
			"",
			func() {
				s.serverFrontendClient.EXPECT().RegisterDomain(gomock.Any(), &types.RegisterDomainRequest{
					Name: "test-domain",
					WorkflowExecutionRetentionPeriodInDays: 3,
					IsGlobalDomain: false,
				}).Return(nil)
			},
		},
		{
			"global",
			"cadence --do test-domain domain register --global_domain true",
			"",
			func() {
				s.serverFrontendClient.EXPECT().RegisterDomain(gomock.Any(), &types.RegisterDomainRequest{
					Name: "test-domain",
					WorkflowExecutionRetentionPeriodInDays: 3,
					IsGlobalDomain: true,
				}).Return(nil)
			},
		},
		{
			"domain with other options",
			"cadence --do test-domain domain register --global_domain true --retention 5 --desc description --domain_data key1=value1,key2=value2 --active_cluster c1 --clusters c1 c2",
			"",
			func() {
				s.serverFrontendClient.EXPECT().RegisterDomain(gomock.Any(), &types.RegisterDomainRequest{
					Name: "test-domain",
					WorkflowExecutionRetentionPeriodInDays: 5,
					IsGlobalDomain: true,
					Description: "description",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					ActiveClusterName: "c1",
					Clusters: []*types.ClusterReplicationConfiguration{
						{
							ClusterName: "c1",
						},
						{
							ClusterName: "c2",
						},
					},
				}).Return(nil)
			},
		},
		{
			"domain exists",
			"cadence --do test-domain domain register --global_domain true",
			"Domain test-domain already registered",
			func() {
				s.serverFrontendClient.EXPECT().RegisterDomain(gomock.Any(), &types.RegisterDomainRequest{
					Name: "test-domain",
					WorkflowExecutionRetentionPeriodInDays: 3,
					IsGlobalDomain: true,
				}).Return(&types.DomainAlreadyExistsError{})
			},
		},
		{
			"failed",
			"cadence --do test-domain domain register --global_domain true",
			"Register Domain operation failed",
			func() {
				s.serverFrontendClient.EXPECT().RegisterDomain(gomock.Any(), &types.RegisterDomainRequest{
					Name: "test-domain",
					WorkflowExecutionRetentionPeriodInDays: 3,
					IsGlobalDomain: true,
				}).Return(&types.BadRequestError{"fake error"})
			},
		},
		{
			"missing flag",
			"cadence domain register",
			"Option domain is required",
			nil,
		},
		{
			"invalid global domain flag",
			"cadence --do test-domain domain register --global_domain invalid",
			"format is invalid",
			nil,
		},
		{
			"invalid history archival status",
			"cadence --do test-domain domain register --global_domain false --history_archival_status invalid",
			"failed to parse",
			nil,
		},
		{
			"invalid visibility archival status",
			"cadence --do test-domain domain register --global_domain false --visibility_archival_status invalid",
			"failed to parse",
			nil,
		},
	})
}

func (s *cliAppSuite) TestDomainUpdate() {
	describeResponse := &types.DescribeDomainResponse{
		DomainInfo: &types.DomainInfo{
			Name: "test-domain",
			Description: "a test domain",
			OwnerEmail: "test@cadence.io",
			Data: map[string]string{
				"key1": "value1",
			},
		},
		Configuration: &types.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: 3,
		},
		ReplicationConfiguration: &types.DomainReplicationConfiguration{
			ActiveClusterName: "c1",
			Clusters: []*types.ClusterReplicationConfiguration{
				{
					ClusterName: "c1",
				},
				{
					ClusterName: "c2",
				},
			},
		},
	}

	s.testcaseHelper([]testcase{
		{
			"update nothing",
			"cadence --do test-domain domain update",
			"",
			func() {
				s.serverFrontendClient.EXPECT().DescribeDomain(gomock.Any(), &types.DescribeDomainRequest{
					Name: common.StringPtr("test-domain"),
				}).Return(describeResponse, nil)

				s.serverFrontendClient.EXPECT().UpdateDomain(gomock.Any(), &types.UpdateDomainRequest{
					Name: "test-domain",
					Description: common.StringPtr("a test domain"),
					OwnerEmail: common.StringPtr("test@cadence.io"),
					Data: nil,
					WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(3),
					EmitMetric: common.BoolPtr(false),
					HistoryArchivalURI: common.StringPtr(""),
					VisibilityArchivalURI: common.StringPtr(""),
					ActiveClusterName: nil,
					Clusters: nil,
					}).Return(&types.UpdateDomainResponse{}, nil)
			},
		},
		{
			"update description",
			"cadence --do test-domain domain update --desc new-description",
			"",
			func ()  {
				s.serverFrontendClient.EXPECT().DescribeDomain(gomock.Any(), &types.DescribeDomainRequest{
					Name: common.StringPtr("test-domain"),
				}).Return(describeResponse, nil)
				s.serverFrontendClient.EXPECT().UpdateDomain(gomock.Any(), &types.UpdateDomainRequest{
					Name: "test-domain",
					Description: common.StringPtr("new-description"),
					OwnerEmail: common.StringPtr("test@cadence.io"),
					Data: nil,
					WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(3),
					EmitMetric: common.BoolPtr(false),
					HistoryArchivalURI: common.StringPtr(""),
					VisibilityArchivalURI: common.StringPtr(""),
					ActiveClusterName: nil,
					Clusters: nil,
					}).Return(&types.UpdateDomainResponse{}, nil)
			},
		},
		{
			"failover",
			"cadence --do test-domain domain update --ac c2",
			"",
			func ()  {
				s.serverFrontendClient.EXPECT().UpdateDomain(gomock.Any(), &types.UpdateDomainRequest{
					Name: "test-domain",
					ActiveClusterName: common.StringPtr("c2"),
					}).Return(&types.UpdateDomainResponse{}, nil)
			},
		},
		{
			"graceful failover",
			"cadence --do test-domain domain update --ac c2 --failover_type grace --failover_timeout_seconds 10",
			"",
			func ()  {
				s.serverFrontendClient.EXPECT().UpdateDomain(gomock.Any(), &types.UpdateDomainRequest{
					Name: "test-domain",
					ActiveClusterName: common.StringPtr("c2"),
					FailoverTimeoutInSeconds: common.Int32Ptr(10),
					}).Return(&types.UpdateDomainResponse{}, nil)
			},
		},
		{
			"domain not exist",
			"cadence --do test-domain domain update --desc new-description",
			"does not exist",
			func ()  {
				s.serverFrontendClient.EXPECT().DescribeDomain(gomock.Any(), &types.DescribeDomainRequest{
					Name: common.StringPtr("test-domain"),
				}).Return(nil, &types.EntityNotExistsError{})
			},
		},
		{
			"describe failure",
			"cadence --do test-domain domain update --desc new-description",
			"describe error",
			func ()  {
				s.serverFrontendClient.EXPECT().DescribeDomain(gomock.Any(), &types.DescribeDomainRequest{
					Name: common.StringPtr("test-domain"),
				}).Return(nil, fmt.Errorf("describe error"))
			},
		},
		{
			"update failure",
			"cadence --do test-domain domain update --desc new-description",
			"update error",
			func ()  {
				s.serverFrontendClient.EXPECT().DescribeDomain(gomock.Any(), &types.DescribeDomainRequest{
					Name: common.StringPtr("test-domain"),
				}).Return(describeResponse, nil)
				s.serverFrontendClient.EXPECT().UpdateDomain(gomock.Any(), &types.UpdateDomainRequest{
					Name: "test-domain",
					Description: common.StringPtr("new-description"),
					OwnerEmail: common.StringPtr("test@cadence.io"),
					Data: nil,
					WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(3),
					EmitMetric: common.BoolPtr(false),
					HistoryArchivalURI: common.StringPtr(""),
					VisibilityArchivalURI: common.StringPtr(""),
					ActiveClusterName: nil,
					Clusters: nil,
					}).Return(nil, fmt.Errorf("update error"))
			},
		},
	})
}

func (s *cliAppSuite) TestListDomains() {
	s.testcaseHelper([]testcase{
		{
			"list domains by default",
			"cadence admin domain list",
			"",
			func() {
				s.serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(&types.ListDomainsResponse{
					Domains: []*types.DescribeDomainResponse{
						{
							DomainInfo: &types.DomainInfo{
								Name: "test-domain",
								Status: types.DomainStatusRegistered.Ptr(),
							},
							ReplicationConfiguration: &types.DomainReplicationConfiguration{},
							Configuration: &types.DomainConfiguration{},
							FailoverInfo: &types.FailoverInfo{},
						},
					},
				}, nil)
			},
		},
	})
}
