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
	"context"
	"errors"
	"path"
	"strings"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
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
	errContextTimeout           = errors.New("activity aborted because context timed out")
	errInvalidGetHistoryRequest = &shared.BadRequestError{Message: "Get archived history request is invalid"}
	errGetHistoryTokenCorrupted = &shared.BadRequestError{Message: "Next page token is corrupted."}
	errHistoryNotExist          = &shared.BadRequestError{Message: "Requested workflow history does not exist."}
)

type (
	// HistoryArchiverConfig configs the filestore implementation of archiver.HistoryArchiver interface
	HistoryArchiverConfig struct {
		*archiver.HistoryIteratorConfig
	}

	historyArchiver struct {
		historyManager   persistence.HistoryManager
		historyV2Manager persistence.HistoryV2Manager
		logger           log.Logger
		config           *HistoryArchiverConfig

		// only set in test code
		historyIterator archiver.HistoryIterator
	}

	getHistoryToken struct {
		CloseFailoverVersion int64
		NextBatchIdx         int
	}
)

func newHistoryArchiver(
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	logger log.Logger,
	config *HistoryArchiverConfig,
	historyIterator archiver.HistoryIterator,
) *historyArchiver {
	return &historyArchiver{
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
	request *archiver.ArchiveHistoryRequest,
	opts ...archiver.ArchiveOption,
) error {
	logger := tagLoggerWithArchiveHistoryRequest(h.logger, request)

	if !h.ValidateURI(URI) {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errInvalidURI), tag.ArchivalURI(URI))
		return archiver.ErrArchiveNonRetriable
	}

	if err := validateArchiveRequest(request); err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errInvalidRequest), tag.Error(err))
		return archiver.ErrArchiveNonRetriable
	}

	historyIterator := h.historyIterator
	if historyIterator == nil {
		var err error
		historyIterator, err = archiver.NewHistoryIterator(request, h.historyManager, h.historyV2Manager, h.config.HistoryIteratorConfig, nil, nil)
		if err != nil {
			// this should not happen
			logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errConstructHistoryIterator), tag.Error(err))
			return archiver.ErrArchiveNonRetriable
		}
	}

	historyBatches := []*shared.History{}
	for historyIterator.HasNext() {
		currHistoryBatches, err := getNextSetofBatches(ctx, historyIterator)
		if err != nil {
			logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errReadHistory), tag.Error(err))
			return archiver.ErrArchiveNonRetriable
		}

		if historyMutated(request, currHistoryBatches, !historyIterator.HasNext()) {
			logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errHistoryMutated))
			return archiver.ErrArchiveNonRetriable
		}

		historyBatches = append(historyBatches, currHistoryBatches...)
	}

	encodedHistoryBatches, err := encodeHistoryBatches(historyBatches)
	if err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errEncodeHistory), tag.Error(err))
		return archiver.ErrArchiveNonRetriable
	}

	dirPath := getDirPathFromURI(URI)
	if err = mkdirAll(dirPath); err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errMakeDirectory), tag.Error(err))
		return archiver.ErrArchiveNonRetriable
	}

	filename := constructFilename(request.DomainID, request.WorkflowID, request.RunID, request.CloseFailoverVersion)
	if err := writeFile(path.Join(dirPath, filename), encodedHistoryBatches); err != nil {
		logger.Error(archiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errWriteFile), tag.Error(err))
		return archiver.ErrArchiveNonRetriable
	}

	return nil
}

func (h *historyArchiver) Get(
	ctx context.Context,
	URI string,
	request *archiver.GetHistoryRequest,
) (*archiver.GetHistoryResponse, error) {
	if !h.ValidateURI(URI) {
		return nil, errors.New(errInvalidURI)
	}

	if err := validateGetRequest(request); err != nil {
		return nil, errInvalidGetHistoryRequest
	}

	dirPath := getDirPathFromURI(URI)
	exists, err := directoryExists(dirPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errHistoryNotExist
	}

	var token *getHistoryToken
	if request.NextPageToken != nil {
		token, err = deserializeGetHistoryToken(request.NextPageToken)
		if err != nil {
			return nil, errGetHistoryTokenCorrupted
		}
	} else if request.CloseFailoverVersion != nil {
		token = &getHistoryToken{
			CloseFailoverVersion: *request.CloseFailoverVersion,
			NextBatchIdx:         0,
		}
	} else {
		filenames, err := listFilesByPrefix(dirPath, constructFilenamePrefix(request.DomainID, request.WorkflowID, request.RunID))
		if err != nil {
			return nil, err
		}
		highestVersion := int64(-1)
		for _, filename := range filenames {
			version, err := extractCloseFailoverVersion(filename)
			if err == nil && version > highestVersion {
				highestVersion = version
			}
		}
		token = &getHistoryToken{
			CloseFailoverVersion: highestVersion,
			NextBatchIdx:         0,
		}
	}

	filename := constructFilename(request.DomainID, request.WorkflowID, request.RunID, token.CloseFailoverVersion)
	filepath := path.Join(dirPath, filename)
	exists, err = fileExists(filepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errHistoryNotExist
	}

	encodedHistoryBatches, err := readFile(filepath)
	if err != nil {
		return nil, err
	}

	historyBatches, err := decodeHistoryBatches(encodedHistoryBatches)
	if err != nil {
		return nil, err
	}
	historyBatches = historyBatches[token.NextBatchIdx:]

	response := &archiver.GetHistoryResponse{}
	numOfEvents := 0
	numOfBatches := 0
	for _, batch := range historyBatches {
		response.HistoryBatches = append(response.HistoryBatches, batch)
		numOfBatches++
		numOfEvents += len(batch.Events)
		if numOfEvents >= request.PageSize {
			break
		}
	}

	if numOfBatches < len(historyBatches) {
		token.NextBatchIdx += numOfBatches
		nextToken, err := serializeGetHistoryToken(token)
		if err != nil {
			return nil, err
		}
		response.NextPageToken = nextToken
	}

	return response, nil
}

func (h *historyArchiver) ValidateURI(URI string) bool {
	if !strings.HasPrefix(URI, fileStoreScheme) {
		return false
	}

	return validateDirPath(getDirPathFromURI(URI))
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
