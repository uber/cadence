package persistence

import (
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/require"
	"github.com/uber-common/bark"
	"testing"
	"sync"
	"github.com/uber/cadence/common"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"time"
)

type (
	historySerializerSuite struct {
		suite.Suite
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
		logger    bark.Logger
	}
)

func TestHistoryBuilderSuite(t *testing.T) {
	s := new(historySerializerSuite)
	suite.Run(t, s)
}

func (s *historySerializerSuite) SetupTest() {
	s.logger = bark.NewLoggerFromLogrus(log.New())
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

func (s *historySerializerSuite) TestSerializerFactory() {

	concurrency := 5
	startWG := sync.WaitGroup{}
	doneWG := sync.WaitGroup{}

	startWG.Add(1)
	doneWG.Add(concurrency)

	factory := NewHistorySerializerFactory()

	event1 := &workflow.HistoryEvent{
		EventId: common.Int64Ptr(999),
		Timestamp: common.Int64Ptr(time.Now().UnixNano()),
		EventType: common.EventTypePtr(workflow.EventType_ActivityTaskCompleted),
		ActivityTaskCompletedEventAttributes: &workflow.ActivityTaskCompletedEventAttributes {
			Result_: []byte("result-1-event-1"),
			ScheduledEventId: common.Int64Ptr(4),
			StartedEventId: common.Int64Ptr(5),
			Identity: common.StringPtr("event-1"),
		},
	}

	for i := 0; i < concurrency; i++ {

		go func() {

			startWG.Wait()
			defer doneWG.Done()

			serializer, err := factory.Get(common.EncodingTypeGob)
			s.NotNil(err)
			s.Equal(ErrUnknownEncodingType, err)

			serializer, err = factory.Get(common.EncodingTypeJSON)
			s.Nil(err)
			s.NotNil(serializer)
			_, ok := serializer.(*jsonHistorySerializer)
			s.True(ok)

			hist := []*workflow.HistoryEvent{event1}
			sh, err := serializer.Serialize(MaxSupportedHistoryVersion+1, hist)
			s.NotNil(err)
			s.Equal(ErrHistoryVersionIncompatible, err)

			sh, err = serializer.Serialize(1, hist)
			s.Nil(err)
			s.NotNil(sh)
			s.Equal(1, sh.Version)
			s.Equal(common.EncodingTypeJSON, sh.EncodingType)

			dh, err := serializer.Deserialize(2, sh.Data)
			s.NotNil(err)
			s.Equal(ErrHistoryVersionIncompatible, err)
			s.Nil(dh)

			dh, err = serializer.Deserialize(1, sh.Data)
			s.Nil(err)
			s.NotNil(dh)

			s.Equal(dh.Version, 1)
			s.Equal(len(dh.Events), 1)
			s.Equal(event1.GetEventId(), dh.Events[0].GetEventId())
			s.Equal(event1.GetTimestamp(), dh.Events[0].GetTimestamp())
			s.Equal(event1.GetEventType(), dh.Events[0].GetEventType())
			s.Equal(event1.GetActivityTaskCompletedEventAttributes().GetResult_(), dh.Events[0].GetActivityTaskCompletedEventAttributes().GetResult_())

		} ()
	}

	startWG.Done()
	succ := common.AwaitWaitGroup(&doneWG, 10*time.Second)
	s.True(succ, "test timed out")
}