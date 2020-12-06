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
