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

package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log/loggerimpl"
)

func TestAckManager(t *testing.T) {
	logger, err := loggerimpl.NewDevelopment()
	assert.Nil(t, err)
	m := NewAckManager(logger)
	m.SetAckLevel(100)
	assert.EqualValues(t, 100, m.GetAckLevel())
	assert.EqualValues(t, 100, m.GetReadLevel())
	const t1 = 200
	const t2 = 220
	const t3 = 320
	const t4 = 340
	const t5 = 360

	err = m.ReadItem(t1)
	assert.Nil(t, err)
	assert.EqualValues(t, 100, m.GetAckLevel())
	assert.EqualValues(t, t1, m.GetReadLevel())

	err = m.ReadItem(t2)
	assert.Nil(t, err)
	assert.EqualValues(t, 100, m.GetAckLevel())
	assert.EqualValues(t, t2, m.GetReadLevel())

	m.AckItem(t2)
	assert.EqualValues(t, 100, m.GetAckLevel())
	assert.EqualValues(t, t2, m.GetReadLevel())

	m.AckItem(t1)
	assert.EqualValues(t, t2, m.GetAckLevel())
	assert.EqualValues(t, t2, m.GetReadLevel())

	m.SetAckLevel(300)
	assert.EqualValues(t, 300, m.GetAckLevel())
	assert.EqualValues(t, 300, m.GetReadLevel())

	err = m.ReadItem(t3)
	assert.Nil(t, err)
	assert.EqualValues(t, 300, m.GetAckLevel())
	assert.EqualValues(t, t3, m.GetReadLevel())

	err = m.ReadItem(t4)
	assert.Nil(t, err)
	assert.EqualValues(t, 300, m.GetAckLevel())
	assert.EqualValues(t, t4, m.GetReadLevel())

	m.AckItem(t3)
	assert.EqualValues(t, t3, m.GetAckLevel())
	assert.EqualValues(t, t4, m.GetReadLevel())

	m.AckItem(t4)
	assert.EqualValues(t, t4, m.GetAckLevel())
	assert.EqualValues(t, t4, m.GetReadLevel())

	m.SetReadLevel(t5)
	assert.EqualValues(t, t5, m.GetReadLevel())
}
