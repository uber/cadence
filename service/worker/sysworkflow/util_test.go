package sysworkflow

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common"
	"testing"
)

type UtilSuite struct {
	*require.Assertions
	suite.Suite
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *UtilSuite) TestNewHistoryBlobKey() {
	testCases := []struct {
		domainID string
		workflowID string
		runID string
		pageToken string
		closeFailoverVersion string
		expectError bool
		expectBuiltKey string
	}{
		{
			domainID: "",
			expectError: true,
		},
		{
			domainID: "testDomainID",
			workflowID: "testWorkflowID",
			runID: "testRunID",
			pageToken: "testPageToken",
			closeFailoverVersion: "testCloseFailoverVersion",
			expectError: false,
			expectBuiltKey: "17971674567288329890367046253745284795510285995943906173973_testPageToken_testCloseFailoverVersion.history",
		},
	}

	for _, tc := range testCases {
		key, err := NewHistoryBlobKey(tc.domainID, tc.workflowID, tc.runID, tc.pageToken, tc.closeFailoverVersion)
		if tc.expectError {
			s.Error(err)
			s.Nil(key)
		} else {
			s.NoError(err)
			s.NotNil(key)
			s.Equal(tc.expectBuiltKey, key.String())
		}
	}
}

func (s *UtilSuite) TestConvertHeaderToTags() {
	testCases := []struct {
		header *HistoryBlobHeader
		expectTags map[string]string
	}{
		{
			header: nil,
			expectTags: map[string]string{},
		},
		{
			header: &HistoryBlobHeader{},
			expectTags: map[string]string{},
		},
		{
			header: &HistoryBlobHeader{
				DomainID: nil,
			},
			expectTags: map[string]string{},
		},
		{
			header: &HistoryBlobHeader{
				DomainID: common.StringPtr("test-domain-id"),
			},
			expectTags: map[string]string{"domain_id": "test-domain-id"},
		},
		{
			header: &HistoryBlobHeader{
				EventCount: nil,
			},
			expectTags: map[string]string{},
		},
		{
			header: &HistoryBlobHeader{
				DomainID: common.StringPtr("test-domain-id"),
				EventCount: common.Int64Ptr(9),
			},
			expectTags: map[string]string{
				"domain_id": "test-domain-id",
				"event_count": "9",
			},
		},
	}

	for _, tc := range testCases {
		tags, err := ConvertHeaderToTags(tc.header)
		s.NoError(err)
		s.Equal(tc.expectTags, tags)
	}
}