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

package filestore

import (
	"errors"

	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

type (
	archiverFactoryImpl struct {
		historyArchiverConfig *HistoryArchiverConfig
	}
)

// NewArchiverFactory returns a new ArchiverFactory for filestore
func NewArchiverFactory(historyArchiverConfig *HistoryArchiverConfig) archiver.Factory {
	return &archiverFactoryImpl{historyArchiverConfig: historyArchiverConfig}
}

func (f *archiverFactoryImpl) NewHistoryArchiver(
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	logger log.Logger,
	metricsClient metrics.Client,
) (archiver.HistoryArchiver, error) {
	historyArchiver := newHistoryArchiver(historyManager, historyV2Manager, logger, f.historyArchiverConfig, nil)
	return historyArchiver, nil
}

func (f *archiverFactoryImpl) NewVisibilityArchiver() (archiver.VisibilityArchiver, error) {
	return nil, errors.New("visibility archiver is not implemented")
}
