// Copyright (c) 2017 Uber Technologies, Inc.
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

package persistencetests

import (
	"context"
	"os"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type (
	// QueuePersistenceSuite contains queue persistence tests
	QueuePersistenceSuite struct {
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

// SetupSuite implementation
func (s *QueuePersistenceSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

// SetupTest implementation
func (s *QueuePersistenceSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

// TearDownSuite implementation
func (s *QueuePersistenceSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

// TestDomainReplicationQueue tests domain replication queue operations
func (s *QueuePersistenceSuite) TestDomainReplicationQueue() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	numMessages := 100
	concurrentSenders := 10
	messageChan := make(chan []byte)

	go func() {
		for i := 0; i < numMessages; i++ {
			messageChan <- []byte{1}
		}
		close(messageChan)
	}()

	wg := sync.WaitGroup{}
	wg.Add(concurrentSenders)

	for i := 0; i < concurrentSenders; i++ {
		go func() {
			defer wg.Done()
			for message := range messageChan {
				err := s.Publish(ctx, message)
				s.Nil(err, "Enqueue message failed.")
			}
		}()
	}

	wg.Wait()

	result, err := s.GetReplicationMessages(ctx, -1, numMessages)
	s.Nil(err, "GetReplicationMessages failed.")
	s.Len(result, numMessages)
}

// TestQueueMetadataOperations tests queue metadata operations
func (s *QueuePersistenceSuite) TestQueueMetadataOperations() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	clusterAckLevels, err := s.GetAckLevels(ctx)
	s.Require().NoError(err)
	s.Assert().Len(clusterAckLevels, 0)

	err = s.UpdateAckLevel(ctx, 10, "test1")
	s.Require().NoError(err)

	clusterAckLevels, err = s.GetAckLevels(ctx)
	s.Require().NoError(err)
	s.Assert().Len(clusterAckLevels, 1)
	s.Assert().Equal(int64(10), clusterAckLevels["test1"])

	err = s.UpdateAckLevel(ctx, 20, "test1")
	s.Require().NoError(err)

	clusterAckLevels, err = s.GetAckLevels(ctx)
	s.Require().NoError(err)
	s.Assert().Len(clusterAckLevels, 1)
	s.Assert().Equal(int64(20), clusterAckLevels["test1"])

	err = s.UpdateAckLevel(ctx, 25, "test2")
	s.Require().NoError(err)

	err = s.UpdateAckLevel(ctx, 24, "test2")
	s.Require().NoError(err)

	clusterAckLevels, err = s.GetAckLevels(ctx)
	s.Require().NoError(err)
	s.Assert().Len(clusterAckLevels, 2)
	s.Assert().Equal(int64(20), clusterAckLevels["test1"])
	s.Assert().Equal(int64(25), clusterAckLevels["test2"])
}

// TestDomainReplicationDLQ tests domain DLQ operations
func (s *QueuePersistenceSuite) TestDomainReplicationDLQ() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	maxMessageID := int64(100)
	numMessages := 100
	concurrentSenders := 10
	messageChan := make(chan []byte)

	go func() {
		for i := 0; i < numMessages; i++ {
			messageChan <- []byte{}
		}
		close(messageChan)
	}()

	wg := sync.WaitGroup{}
	wg.Add(concurrentSenders)

	for i := 0; i < concurrentSenders; i++ {
		go func() {
			defer wg.Done()
			for message := range messageChan {
				err := s.PublishToDomainDLQ(ctx, message)
				s.Nil(err, "Enqueue message failed.")
			}
		}()
	}

	wg.Wait()

	result1, token, err := s.GetMessagesFromDomainDLQ(ctx, -1, maxMessageID, numMessages/2, nil)
	s.Nil(err, "GetReplicationMessages failed.")
	s.NotNil(token)
	result2, token, err := s.GetMessagesFromDomainDLQ(ctx, -1, maxMessageID, numMessages, token)
	s.Nil(err, "GetReplicationMessages failed.")
	s.Equal(len(token), 0)
	s.Equal(len(result1)+len(result2), numMessages)
	_, _, err = s.GetMessagesFromDomainDLQ(ctx, -1, 1<<63-1, numMessages, nil)
	s.NoError(err, "GetReplicationMessages failed.")
	s.Equal(len(token), 0)

	size, err := s.GetDomainDLQSize(ctx)
	s.NoError(err, "GetDomainDLQSize failed")
	s.Equal(int64(numMessages), size)

	lastMessageID := result2[len(result2)-1].ID
	err = s.DeleteMessageFromDomainDLQ(ctx, lastMessageID)
	s.NoError(err)
	result3, token, err := s.GetMessagesFromDomainDLQ(ctx, -1, maxMessageID, numMessages, token)
	s.Nil(err, "GetReplicationMessages failed.")
	s.Equal(len(token), 0)
	s.Equal(len(result3), numMessages-1)

	err = s.RangeDeleteMessagesFromDomainDLQ(ctx, -1, lastMessageID)
	s.NoError(err)
	result4, token, err := s.GetMessagesFromDomainDLQ(ctx, -1, maxMessageID, numMessages, token)
	s.Nil(err, "GetReplicationMessages failed.")
	s.Equal(len(token), 0)
	s.Equal(len(result4), 0)
}

// TestDomainDLQMetadataOperations tests queue metadata operations
func (s *QueuePersistenceSuite) TestDomainDLQMetadataOperations() {
	clusterName := "test"
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	ackLevel, err := s.GetDomainDLQAckLevel(ctx)
	s.Require().NoError(err)
	s.Equal(0, len(ackLevel))

	err = s.UpdateDomainDLQAckLevel(ctx, 10, clusterName)
	s.NoError(err)

	ackLevel, err = s.GetDomainDLQAckLevel(ctx)
	s.Require().NoError(err)
	s.Equal(int64(10), ackLevel[clusterName])

	err = s.UpdateDomainDLQAckLevel(ctx, 1, clusterName)
	s.NoError(err)

	ackLevel, err = s.GetDomainDLQAckLevel(ctx)
	s.Require().NoError(err)
	s.Equal(int64(10), ackLevel[clusterName])
}
