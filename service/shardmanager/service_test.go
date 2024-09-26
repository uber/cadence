package shardmanager

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/resource"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	resourceMock := resource.NewMockResource(ctrl)
	factoryMock := resource.NewMockResourceFactory(ctrl)
	factoryMock.EXPECT().NewResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(resourceMock, nil).AnyTimes()

	service, err := NewService(&resource.Params{}, factoryMock)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestServiceStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceMock := resource.NewMockResource(ctrl)
	resourceMock.EXPECT().GetLogger().Return(log.NewNoop()).AnyTimes()
	resourceMock.EXPECT().Start().Return()
	resourceMock.EXPECT().Stop().Return()

	service := &Service{
		Resource: resourceMock,
		status:   common.DaemonStatusInitialized,
		stopC:    make(chan struct{}),
	}

	go service.Start()

	time.Sleep(100 * time.Millisecond) // The code assumes the service is started when calling Stop
	assert.Equal(t, int32(common.DaemonStatusStarted), atomic.LoadInt32(&service.status))

	service.Stop()
	assert.Equal(t, int32(common.DaemonStatusStopped), atomic.LoadInt32(&service.status))
}

func TestStartAndStopReturnsImmediatelyWhenAlreadyStopped(t *testing.T) {

	service := &Service{
		status: common.DaemonStatusStopped,
	}

	service.Start()
	assert.Equal(t, int32(common.DaemonStatusStopped), atomic.LoadInt32(&service.status))

	service.Stop()
	assert.Equal(t, int32(common.DaemonStatusStopped), atomic.LoadInt32(&service.status))
}
