// Copyright (c) 2020 Uber Technologies, Inc.
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

package gcloud

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

func (s *utilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}
func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(utilSuite))
}

type utilSuite struct {
	*require.Assertions
	suite.Suite
}

func (s *utilSuite) TestEncodeDecodeHistoryBatches() {
	historyBatches := []*shared.History{
		&shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventId: common.Int64Ptr(common.FirstEventID),
					Version: common.Int64Ptr(1),
				},
			},
		},
		&shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventId:   common.Int64Ptr(common.FirstEventID + 1),
					Timestamp: common.Int64Ptr(time.Now().UnixNano()),
					Version:   common.Int64Ptr(1),
				},
				&shared.HistoryEvent{
					EventId: common.Int64Ptr(common.FirstEventID + 2),
					Version: common.Int64Ptr(2),
					DecisionTaskStartedEventAttributes: &shared.DecisionTaskStartedEventAttributes{
						Identity: common.StringPtr("some random identity"),
					},
				},
			},
		},
	}

	encodedHistoryBatches, err := encode(historyBatches)
	s.NoError(err)

	decodedHistoryBatches, err := decodeHistoryBatches(encodedHistoryBatches)
	s.NoError(err)
	s.Equal(historyBatches, decodedHistoryBatches)
}

func (s *utilSuite) TestconstructHistoryFilename() {
	testCases := []struct {
		domainID             string
		workflowID           string
		runID                string
		closeFailoverVersion int64
		expectBuiltName      string
	}{
		{
			domainID:             "testDomainID",
			workflowID:           "testWorkflowID",
			runID:                "testRunID",
			closeFailoverVersion: 5,
			expectBuiltName:      "17971674567288329890367046253745284795510285995943906173973_5_0.history",
		},
	}

	for _, tc := range testCases {
		filename := constructHistoryFilenameMultipart(tc.domainID, tc.workflowID, tc.runID, tc.closeFailoverVersion, 0)
		s.Equal(tc.expectBuiltName, filename)
	}
}

func (s *utilSuite) TestSerializeDeserializeGetHistoryToken() {
	token := &getHistoryToken{
		CloseFailoverVersion: 101,
		BatchIdxOffset:       20,
	}

	serializedToken, err := serializeToken(token)
	s.Nil(err)

	deserializedToken, err := deserializeGetHistoryToken(serializedToken)
	s.Nil(err)
	s.Equal(token, deserializedToken)
}

func (s *utilSuite) TestConstructHistoryFilenamePrefix() {
	s.Equal("28646288347718592068344541402884576509131521284625246243", constructHistoryFilenamePrefix("domainID", "workflowID", "runID"))
}

func (s *utilSuite) TestConstructHistoryFilenameMultipart() {
	s.Equal("28646288347718592068344541402884576509131521284625246243_-24_0.history", constructHistoryFilenameMultipart("domainID", "workflowID", "runID", -24, 0))
}

func (s *utilSuite) TestConstructVisibilityFilenamePrefix() {
	s.Equal("domainID/startTimeout", constructVisibilityFilenamePrefix("domainID", indexKeyStartTimeout))
}

func (s *utilSuite) TestConstructTimeBasedSearchKey() {
	s.Equal("domainID/startTimeout_1970-01-01T", constructTimeBasedSearchKey("domainID", indexKeyStartTimeout, 1580819141, "Day"))
}

func (s *utilSuite) TestConstructVisibilityFilename() {
	s.Equal("domainID/startTimeout_1970-01-01T00:24:32Z_4346151385925082125_8344541402884576509_131521284625246243.visibility", constructVisibilityFilename("domainID", "workflowTypeName", "workflowID", "runID", indexKeyStartTimeout, 1472313624305))
}

func (s *utilSuite) TestWorkflowIdPrecondition() {
	testCases := []struct {
		workflowID     string
		fileName       string
		expectedResult bool
	}{
		{
			workflowID:     "testWorkflowID",
			fileName:       "startTimeout_1970-01-01T_testWorkflowID_workflowTypeName_runID.visibility",
			expectedResult: true,
		},
		{
			workflowID:     "testWorkflowID",
			fileName:       "startTimeout_1970-01-01T_UnkownWorkflowID_workflowTypeName_runID.visibility",
			expectedResult: false,
		},
		{
			workflowID:     "",
			fileName:       "startTimeout_1970-01-01T_UnkownWorkflowID_workflowTypeName_runID.visibility",
			expectedResult: true,
		},
	}

	precondition := existWorkflowID()
	for _, testCase := range testCases {
		s.Equal(precondition(testCase.fileName, testCase.workflowID), testCase.expectedResult)
	}

}

func (s *utilSuite) TestRunIdPrecondition() {
	testCases := []struct {
		workflowID     string
		runID          string
		fileName       string
		expectedResult bool
	}{
		{
			workflowID:     "testWorkflowID",
			runID:          "runID",
			fileName:       "startTimeout_1970-01-01T_testWorkflowID_workflowTypeName_runID.visibility",
			expectedResult: true,
		},
		{
			workflowID:     "testWorkflowID",
			runID:          "runID",
			fileName:       "startTimeout_1970-01-01T_UnkownWorkflowID_workflowTypeName_runUnKnownID.visibility",
			expectedResult: false,
		},
		{
			workflowID:     "testWorkflowID",
			runID:          "",
			fileName:       "startTimeout_1970-01-01T_UnkownWorkflowID_workflowTypeName_runID.visibility",
			expectedResult: true,
		},
	}

	precondition := existRunID()
	for _, testCase := range testCases {
		s.Equal(precondition(testCase.fileName, testCase.workflowID, testCase.runID), testCase.expectedResult)
	}

}

func (s *utilSuite) TestWorkflowTypeNamePrecondition() {
	testCases := []struct {
		workflowID       string
		runID            string
		workflowTypeName string
		fileName         string
		expectedResult   bool
	}{
		{
			workflowID:       "testWorkflowID",
			runID:            "runID",
			workflowTypeName: "workflowTypeName",
			fileName:         "startTimeout_1970-01-01T_testWorkflowID_workflowTypeName_runID.visibility",
			expectedResult:   true,
		},
		{
			workflowID:       "testWorkflowID",
			runID:            "runID",
			workflowTypeName: "workflowTypeName",
			fileName:         "startTimeout_1970-01-01T_UnkownWorkflowID_workflowUnkownTypeName_runID.visibility",
			expectedResult:   false,
		},
		{
			workflowID:       "testWorkflowID",
			runID:            "runID",
			workflowTypeName: "",
			fileName:         "startTimeout_1970-01-01T_UnkownWorkflowID_workflowTypeName_runID.visibility",
			expectedResult:   true,
		},
	}

	precondition := existWorkflowTypeName()
	for _, testCase := range testCases {
		s.Equal(precondition(testCase.fileName, testCase.workflowID, testCase.runID, testCase.workflowTypeName), testCase.expectedResult)
	}

}
