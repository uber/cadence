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

package gcloud

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"

	"github.com/uber/cadence/common/archiver/gcloud/connector"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/config"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

var (
	errUploadNonRetriable = errors.New("upload non-retriable error")
	errUploadRetriable    = errors.New("upload retriable error")
)

const (
	// URIScheme is the scheme for the gcloud storage implementation
	URIScheme = "gs"

	targetHistoryBlobSize = 2 * 1024 * 1024 // 2MB
	errEncodeHistory      = "failed to encode history batches"
	errBucketHistory      = "failed to get google storage bucket handle"
	errWriteFile          = "failed to write history to google storage"
)

type historyArchiverData struct {
	container     *archiver.HistoryBootstrapContainer
	gcloudStorage connector.Client

	// only set in test code
	historyIterator archiver.HistoryIterator
}

type historyIteratorState struct {
	NextEventID       int64
	FinishedIteration bool
}

type getHistoryToken struct {
	CloseFailoverVersion int64
	FailoverParts        []int64
	NextBatchIdx         int
}

// NewHistoryArchiver creates a new gcloud storage HistoryArchiver
func NewHistoryArchiver(
	container *archiver.HistoryBootstrapContainer,
	config *config.GstorageArchiver,
) (archiver.HistoryArchiver, error) {
	storage, err := connector.NewClient(context.Background(), config)
	if err == nil {
		return newHistoryArchiver(container, nil, storage), nil
	}
	return nil, err
}

func newHistoryArchiver(container *archiver.HistoryBootstrapContainer, historyIterator archiver.HistoryIterator, storage connector.Client) archiver.HistoryArchiver {
	return &historyArchiverData{
		container:       container,
		gcloudStorage:   storage,
		historyIterator: historyIterator,
	}
}

// Archive is used to archive a workflow history. When the context expires the method should stop trying to archive.
// Implementors are free to archive however they want, including implementing retries of sub-operations. The URI defines
// the resource that histories should be archived into. The implementor gets to determine how to interpret the URI.
// The Archive method may or may not be automatically retried by the caller. The ArchiveOptions are used
// to interact with these retries including giving the implementor the ability to cancel retries and record progress
// between retry attempts.
// This method will be invoked after a workflow passes its retention period.
func (h *historyArchiverData) Archive(ctx context.Context, URI archiver.URI, request *archiver.ArchiveHistoryRequest, opts ...archiver.ArchiveOption) (err error) {
	scope := h.container.MetricsClient.Scope(metrics.HistoryArchiverScope, metrics.DomainTag(request.DomainName))
	featureCatalog := archiver.GetFeatureCatalog(opts...)
	sw := scope.StartTimer(metrics.CadenceLatency)
	defer func() {
		sw.Stop()
		if err != nil {
			if err.Error() == errUploadNonRetriable.Error() {
				scope.IncCounter(metrics.ArchiverNonRetryableErrorCount)
			}
		}
	}()

	logger := archiver.TagLoggerWithArchiveHistoryRequestAndURI(h.container.Logger, request, URI.String())

	if err := h.validateURI(ctx, URI); err != nil {
		logger.Error(archiver.ArchiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(archiver.ErrReasonInvalidURI), tag.Error(err))
		return errUploadNonRetriable
	}

	if err := archiver.ValidateHistoryArchiveRequest(request); err != nil {
		logger.Error(archiver.ArchiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(archiver.ErrReasonInvalidArchiveRequest), tag.Error(err))
		return errUploadNonRetriable
	}

	historyIterator := h.historyIterator
	if historyIterator == nil { // will only be set by testing code
		historyIterator, _ = loadHistoryIterator(ctx, request, h.container.HistoryV2Manager, featureCatalog)
	}

	for historyIterator.HasNext() {
		part := getCurrentHistoryPartNumber(historyIterator)
		historyBlob, err := getNextHistoryBlob(ctx, historyIterator)

		if err != nil {
			logger := logger.WithTags(tag.ArchivalArchiveFailReason(archiver.ErrReasonReadHistory), tag.Error(err))
			if !common.IsPersistenceTransientError(err) {
				logger.Error(archiver.ArchiveNonRetriableErrorMsg)
			} else {
				logger.Error(archiver.ArchiveTransientErrorMsg)
			}
			return errUploadNonRetriable
		}

		if historyMutated(request, historyBlob.Body, *historyBlob.Header.IsLast) {
			logger.Error(archiver.ArchiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(archiver.ErrReasonHistoryMutated))
			return archiver.ErrHistoryMutated
		}

		encodedHistoryPart, err := encode(historyBlob.Body)
		if err != nil {
			logger.Error(archiver.ArchiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason(errEncodeHistory), tag.Error(err))
			return errUploadNonRetriable
		}

		filename := constructHistoryFilenameMultipart(request.DomainID, request.WorkflowID, request.RunID, request.CloseFailoverVersion, part)
		if exist, _ := h.gcloudStorage.Exist(ctx, URI, filename); !exist {
			if err := h.gcloudStorage.Upload(ctx, URI, filename, encodedHistoryPart); err != nil {
				logger.Error(archiver.ArchiveTransientErrorMsg, tag.ArchivalArchiveFailReason(errWriteFile), tag.Error(err))
				return errUploadRetriable
			}

			scope.AddCounter(metrics.HistoryArchiverTotalUploadSize, int64(binary.Size(encodedHistoryPart)))
			scope.AddCounter(metrics.HistoryArchiverHistorySize, int64(binary.Size(encodedHistoryPart)))
		}

		saveHistoryIteratorState(ctx, featureCatalog, historyIterator)

	}

	scope.IncCounter(metrics.HistoryArchiverArchiveSuccessCount)
	return
}

// Get is used to access an archived history. When context expires method should stop trying to fetch history.
// The URI identifies the resource from which history should be accessed and it is up to the implementor to interpret this URI.
// This method should thrift errors - see filestore as an example.
func (h *historyArchiverData) Get(ctx context.Context, URI archiver.URI, request *archiver.GetHistoryRequest) (tmp *archiver.GetHistoryResponse, err error) {

	if err := h.validateURI(ctx, URI); err != nil {
		return nil, &shared.BadRequestError{Message: archiver.ErrInvalidURI.Error()}
	}

	if err := archiver.ValidateGetRequest(request); err != nil {
		return nil, &shared.BadRequestError{Message: archiver.ErrInvalidGetHistoryRequest.Error()}
	}

	var token *getHistoryToken
	if request.NextPageToken != nil {
		token, err = deserializeGetHistoryToken(request.NextPageToken)
		if err != nil {
			return nil, &shared.BadRequestError{Message: archiver.ErrNextPageTokenCorrupted.Error()}
		}
	} else if request.CloseFailoverVersion != nil {
		token = &getHistoryToken{
			CloseFailoverVersion: *request.CloseFailoverVersion,
			NextBatchIdx:         0,
		}
	} else {
		highestVersion, failoverParts, err := h.getHighestVersion(ctx, URI, request)
		if err != nil {
			return nil, &shared.InternalServiceError{Message: err.Error()}
		}
		token = &getHistoryToken{
			CloseFailoverVersion: *highestVersion,
			FailoverParts:        failoverParts,
			NextBatchIdx:         0,
		}
	}

	historyBatches := []*shared.History{}
	for _, partNum := range token.FailoverParts {
		filename := constructHistoryFilenameMultipart(request.DomainID, request.WorkflowID, request.RunID, token.CloseFailoverVersion, partNum)
		encodedHistoryBatches, err := h.gcloudStorage.Get(ctx, URI, filename)

		if err != nil {
			return nil, &shared.InternalServiceError{Message: err.Error()}
		}

		if encodedHistoryBatches == nil {
			return nil, &shared.InternalServiceError{Message: "Fail retrieving history file: " + URI.String() + "/" + filename}
		}

		batches, err := decodeHistoryBatches(encodedHistoryBatches)
		if err != nil {
			return nil, &shared.InternalServiceError{Message: err.Error()}
		}
		historyBatches = append(historyBatches, batches...)
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
		nextToken, err := serializeToken(token)
		if err != nil {
			return nil, &shared.InternalServiceError{Message: err.Error()}
		}
		response.NextPageToken = nextToken
	}

	return response, nil

}

// ValidateURI is used to define what a valid URI for an implementation is.
func (h *historyArchiverData) ValidateURI(URI archiver.URI) (err error) {
	if URI.Scheme() != URIScheme {
		return archiver.ErrURISchemeMismatch
	}

	if URI.Path() == "" || URI.Hostname() == "" {
		return archiver.ErrInvalidURI
	}

	return
}

func (h *historyArchiverData) validateURI(ctx context.Context, URI archiver.URI) (err error) {
	if err = h.ValidateURI(URI); err == nil {
		_, err = h.gcloudStorage.Exist(ctx, URI, "")
	}

	return
}

func getNextHistoryBlob(ctx context.Context, historyIterator archiver.HistoryIterator) (*archiver.HistoryBlob, error) {
	historyBlob, err := historyIterator.Next()
	op := func() error {
		historyBlob, err = historyIterator.Next()
		return err
	}
	for err != nil {
		if !common.IsPersistenceTransientError(err) {
			return nil, err
		}
		if contextExpired(ctx) {
			return nil, archiver.ErrContextTimeout
		}
		err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return historyBlob, nil
}

func historyMutated(request *archiver.ArchiveHistoryRequest, historyBatches []*shared.History, isLast bool) bool {
	lastBatch := historyBatches[len(historyBatches)-1].Events
	lastEvent := lastBatch[len(lastBatch)-1]
	lastFailoverVersion := lastEvent.GetVersion()
	if lastFailoverVersion > request.CloseFailoverVersion {
		return true
	}

	if !isLast {
		return false
	}
	lastEventID := lastEvent.GetEventId()
	return lastFailoverVersion != request.CloseFailoverVersion || lastEventID+1 != request.NextEventID
}

func (h *historyArchiverData) getHighestVersion(ctx context.Context, URI archiver.URI, request *archiver.GetHistoryRequest) (*int64, []int64, error) {

	filenames, err := h.gcloudStorage.Query(ctx, URI, constructHistoryFilenamePrefix(request.DomainID, request.WorkflowID, request.RunID))

	if err != nil {
		return nil, nil, err
	}

	var highestVersion *int64
	failoverPart := make([]int64, 0, 0)
	for _, filename := range filenames {
		version, partVersion, err := extractCloseFailoverVersion(filepath.Base(filename))
		if err != nil {
			continue
		}
		if highestVersion == nil || version > *highestVersion {
			highestVersion = &version
		}

		failoverPart = append(failoverPart, partVersion)

	}
	if highestVersion == nil || len(failoverPart) == 0 {
		return nil, nil, archiver.ErrHistoryNotExist
	}

	sort.Slice(failoverPart, func(i, j int) bool { return failoverPart[i] < failoverPart[j] })
	return highestVersion, failoverPart, nil
}

func loadHistoryIterator(ctx context.Context, request *archiver.ArchiveHistoryRequest, historyManager persistence.HistoryManager, featureCatalog *archiver.ArchiveFeatureCatalog) (historyIterator archiver.HistoryIterator, err error) {

	defer func() {
		if err != nil || historyIterator == nil {
			historyIterator, err = archiver.NewHistoryIteratorFromState(request, historyManager, targetHistoryBlobSize, nil)
		}
	}()

	if featureCatalog.ProgressManager != nil {
		historyProgressManager := new(historyIteratorState)
		err = featureCatalog.ProgressManager.LoadProgress(ctx, &historyProgressManager)
		if err == nil {
			iteratorState, _ := json.Marshal(historyProgressManager)
			historyIterator, err = archiver.NewHistoryIteratorFromState(request, historyManager, targetHistoryBlobSize, iteratorState)
		}
	}
	return
}

func saveHistoryIteratorState(ctx context.Context, featureCatalog *archiver.ArchiveFeatureCatalog, historyIterator archiver.HistoryIterator) (err error) {
	var state []byte
	if featureCatalog.ProgressManager != nil {
		state, err = historyIterator.GetState()
		if err == nil {
			err = featureCatalog.ProgressManager.RecordProgress(ctx, state)
		}
	}

	return err
}

func getCurrentHistoryPartNumber(historyIterator archiver.HistoryIterator) int64 {
	historyProgressManager := new(historyIteratorState)
	state, _ := historyIterator.GetState()
	json.Unmarshal(state, &historyProgressManager)
	return historyProgressManager.NextEventID
}
