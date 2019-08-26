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

package collection

const numPriorities = 2

// ChannelPriorityQueue is a priority queue built using channels
type ChannelPriorityQueue struct {
	channels   []chan interface{}
	shutdownCh chan struct{}
}

// NewChannelPriorityQueue returns a ChannelPriorityQueue
func NewChannelPriorityQueue(queueSize int) *ChannelPriorityQueue {
	channels := make([]chan interface{}, numPriorities)
	for i := range channels {
		channels[i] = make(chan interface{}, queueSize)
	}
	return &ChannelPriorityQueue{
		channels:   channels,
		shutdownCh: make(chan struct{}, 1),
	}
}

// Add adds an item to a channel in the queue. This is blocking and waits for
// the queue to get empty if it is full
func (c *ChannelPriorityQueue) Add(priority int, item interface{}) {
	if priority >= numPriorities {
		return
	}
	c.channels[priority] <- item
}

// Remove removes an item from the priority queue. This is blocking till an
// element becomes available in the priority queue
func (c *ChannelPriorityQueue) Remove() interface{} {
	// pick from highest priority if exists
	select {
	case task := <-c.channels[0]:
		return task
	default:
	}

	// blocking select from all priorities
	var task interface{}
	select {
	case task = <-c.channels[0]:
	case task = <-c.channels[1]:
	}
	return task
}

// Close - destroys the channel priority queue
func (c *ChannelPriorityQueue) Close() {
	for _, channel := range c.channels {
		close(channel)
	}
}
