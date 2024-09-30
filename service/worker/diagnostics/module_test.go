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

package diagnostics

import (
	"github.com/Shopify/sarama/mocks"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/messaging/kafka"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/shared"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
)

func Test__Start(t *testing.T) {
	dwTest, mockResource := setuptest(t)
	err := dwTest.Start()
	require.NoError(t, err)
	dwTest.Stop()
	mockResource.Finish(t)
}

func setuptest(t *testing.T) (DiagnosticsWorkflow, *resource.Test) {
	ctrl := gomock.NewController(t)
	mockClientBean := client.NewMockBean(ctrl)
	mockResource := resource.NewTest(t, ctrl, metrics.Worker)
	logger := testlogger.New(t)
	messaginClientMock := messaging.NewMockClient(ctrl)
	messaginClientMock.EXPECT().NewProducer(gomock.Any()).Return(kafka.NewKafkaProducer("test-topic", mocks.NewSyncProducer(t, nil), logger), nil).MinTimes(1)
	sdkClient := mockResource.GetSDKClient()
	mockResource.SDKClient.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.DescribeDomainResponse{}, nil).AnyTimes()
	mockResource.SDKClient.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.PollForDecisionTaskResponse{}, nil).AnyTimes()
	mockResource.SDKClient.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.PollForActivityTaskResponse{}, nil).AnyTimes()
	return New(Params{
		ServiceClient:   sdkClient,
		ClientBean:      mockClientBean,
		MetricsClient:   nil,
		TallyScope:      tally.TestScope(nil),
		MessagingClient: messaginClientMock,
	}), mockResource
}
