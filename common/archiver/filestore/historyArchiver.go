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
	"github.com/uber/cadence/common/backoff"
	"context"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

const (
	fileStoreScheme = "file://"

	archiveNonRetriableErrorMsg = "Archive method encountered an non-retriable error."

	errInvalidURI               = "archive URI is invalid"
	errInvalidRequest           = "archive request is invalid"
	errConstructHistoryIterator = "failed to construct history iterator"
	errReadHistory              = "failed to read history batches"
	errHistoryMutated           = "history was mutated"
	errEncodeHistory            = "failed to encode history batches"
	errMakeDirectory            = "failed to make directory"
	errWriteFile                = "failed to write history to file"
)

var (
	errContextTimeout = errors.New("activity aborted because context timed out")
)

type (
	Config struct {
		*archiver.HistoryIteratorConfig

		StoreDirectory string
	}

	historyArchiver struct {
		clusterMetadata  cluster.Metadata
		historyManager   persistence.HistoryManager
		historyV2Manager persistence.HistoryV2Manager
		logger           log.Logger
		config           *Config

		// only set in test code
		historyIterator archiver.HistoryIterator
	}
)

func NewHistoryArchiver(
	clusterMetadata cluster.Metadata,
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	metricsClient metrics.Client,
	logger log.Logger,
	config *Config,
) archiver.HistoryArchiver {
	return newHistoryArchiver(clusterMetadata, historyManager, historyV2Manager,
		logger, config, nil)
}

func newHistoryArchiver(
	clusterMetadata cluster.Metadata,
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	logger log.Logger,
	config *Config,
	historyIterator archiver.HistoryIterator,
) *historyArchiver {
	return &historyArchiver{
		clusterMetadata:  clusterMetadata,
		historyManager:   historyManager,
		historyV2Manager: historyV2Manager,
		logger:           logger,
		config:           config,
		historyIterator:  historyIterator,
	}
}

func (h *historyArchiver) Archive(
	ctx context.Context,
	URI string,
	request archiver.ArchiveHistoryRequest,
	opts ...archiver.ArchiveOption,
) error {
	logger := tagLoggerWithArchiveHistoryRequest(h.logger, &request)

	if !h.ValidateURI(URI) {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errInvalidURI), tag.ArchivalURI(URI))
		return archiver.ArchiveNonRetriableErr
	}

	historyIterator := h.historyIterator
	if historyIterator == nil {
		var err error
		historyIterator, err = archiver.NewHistoryIterator(request, h.historyManager, h.historyV2Manager, h.config.HistoryIteratorConfig, nil, nil)
		if err != nil {
			logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errConstructHistoryIterator), tag.Error(err))
			return archiver.ArchiveNonRetriableErr
		}
	}

	historyBatches := []*shared.History{}
	for historyIterator.HasNext() {
		currHistoryBatches, err := getNextSetofBatches(ctx, historyIterator)
		if err != nil {
			logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errReadHistory), tag.Error(err))
			return archiver.ArchiveNonRetriableErr
		}

		if historyMutated(&request, currHistoryBatches, !historyIterator.HasNext()) {
			logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errHistoryMutated))
			return archiver.ArchiveNonRetriableErr
		}

		historyBatches = append(historyBatches, currHistoryBatches...)
	}

	encodedHistoryBatches, err := encodeHistoryBatches(historyBatches)
	if err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errEncodeHistory), tag.Error(err))
		return archiver.ArchiveNonRetriableErr
	}

	dirPath := getDirPathFromURI(URI)
	if err = mkdirAll(dirPath); err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errMakeDirectory), tag.Error(err))
		return archiver.ArchiveNonRetriableErr
	}

	filename := constructFilename(&request)
	if err := writeFile(path.Join(dirPath, filename), encodedHistoryBatches); err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errWriteFile), tag.Error(err))
		return archiver.ArchiveNonRetriableErr
	}

	return nil
}

func (h *historyArchiver) Get(
	ctx context.Context,
	URI string,
	request archiver.GetHistoryRequest,
) (archiver.GetHistoryResponse, error) {
	return archiver.GetHistoryResponse{}, nil
}

func (h *historyArchiver) ValidateURI(URI string) bool {
	if !strings.HasPrefix(URI, fileStoreScheme) {
		return false
	}

	dirPath := getDirPathFromURI(URI)
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return false
	}
	return info.IsDir()
}

func getNextSetofBatches(ctx context.Context, historyIterator archiver.HistoryIterator) ([]*shared.History, error) {
	historyBatches, err := historyIterator.Next()
	op := func() error {
		historyBatches, err = historyIterator.Next()
		return err
	}
	for err != nil {
		if !common.IsPersistenceTransientError(err) {
			return nil, err
		}
		if contextExpired(ctx) {
			return nil, errContextTimeout
		}
		err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return historyBatches, nil
}
