package thrift

import (
	"time"

	"github.com/uber/cadence/common"
)

func timeToNano(t *time.Time) *int64 {
	if t == nil {
		return nil
	}

	return common.Int64Ptr(t.UnixNano())
}
