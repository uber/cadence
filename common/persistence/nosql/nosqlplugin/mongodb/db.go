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

package mongodb

import (
	"errors"
	"fmt"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

const (
	// PluginName is the name of the plugin
	PluginName = "mongodb"
)

var (
	errConditionFailed = errors.New("internal condition fail error")
)

// mdb represents a logical connection to MongoDB database
type mdb struct {
	logger log.Logger
}

var _ nosqlplugin.DB = (*mdb)(nil)

// NewMongoDB return a new DB
func NewMongoDB(cfg config.NoSQL, logger log.Logger) (nosqlplugin.DB, error) {
	return nil, fmt.Errorf("TODO")
}

func (db *mdb) Close() {
	panic("TODO")
}

func (db *mdb) PluginName() string {
	return PluginName
}

func (db *mdb) IsNotFoundError(err error) bool {
	panic("TODO")
}

func (db *mdb) IsTimeoutError(err error) bool {
	panic("TODO")
}

func (db *mdb) IsThrottlingError(err error) bool {
	panic("TODO")
}

func (db *mdb) IsConditionFailedError(err error) bool {
	if err == errConditionFailed {
		return true
	}
	return false
}
