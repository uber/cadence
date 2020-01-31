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

package gcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/uber/cadence/common"

	"github.com/dgryski/go-farm"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/archiver"
)

func encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decodeHistoryBatches(data []byte) ([]*shared.History, error) {
	historyBatches := []*shared.History{}
	err := json.Unmarshal(data, &historyBatches)
	if err != nil {
		return nil, err
	}
	return historyBatches, nil
}

func constructHistoryFilename(domainID, workflowID, runID string, version int64) string {
	combinedHash := constructHistoryFilenamePrefix(domainID, workflowID, runID)
	return fmt.Sprintf("%s_%v.history", combinedHash, version)
}

func constructHistoryFilenameMultipart(domainID, workflowID, runID string, version int64, partNumber int) string {
	combinedHash := constructHistoryFilenamePrefix(domainID, workflowID, runID)
	return fmt.Sprintf("%s_%v_%v.history", combinedHash, version, partNumber)
}

func constructHistoryFilenamePrefix(domainID, workflowID, runID string) string {
	return strings.Join([]string{hash(domainID), hash(workflowID), hash(runID)}, "")
}

func hashVisibilityFilenamePrefix(domainID, ID string) string {
	return strings.Join([]string{hash(domainID), hash(ID)}, "")
}

func constructVisibilityFilenamePrefix(domainID, runID, workflowID string) string {
	var combinedDomainRunIDHash string
	if runID != "" {
		combinedDomainRunIDHash = hashVisibilityFilenamePrefix(domainID, runID)
	}

	combinedDomainWorkflowIDHash := hashVisibilityFilenamePrefix(domainID, workflowID)
	return fmt.Sprintf("%s_%s", combinedDomainWorkflowIDHash, combinedDomainRunIDHash)
}

func hash(s string) string {
	return fmt.Sprintf("%v", farm.Fingerprint64([]byte(s)))
}

func contextExpired(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func deserializeGetHistoryToken(bytes []byte) (*getHistoryToken, error) {
	token := &getHistoryToken{}
	err := json.Unmarshal(bytes, token)
	return token, err
}

func extractCloseFailoverVersion(filename string) (int64, int, error) {
	filenameParts := strings.FieldsFunc(filename, func(r rune) bool {
		return r == '_' || r == '.'
	})
	if len(filenameParts) != 4 {
		return -1, 0, errors.New("unknown filename structure")
	}

	failoverVersion, err := strconv.ParseInt(filenameParts[1], 10, 64)
	if err != nil {
		return -1, 0, err
	}

	highestPart, err := strconv.Atoi(filenameParts[2])
	return failoverVersion, highestPart, err
}

func serializeToken(token interface{}) ([]byte, error) {
	if token == nil {
		return nil, nil
	}
	return json.Marshal(token)
}

func decodeVisibilityRecord(data []byte) (*visibilityRecord, error) {
	record := &visibilityRecord{}
	err := json.Unmarshal(data, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func constructVisibilityFilename(domainID, workflowID, runID string, timestamp int64) string {
	t := time.Unix(0, timestamp).In(time.UTC)
	prefix := constructVisibilityFilenamePrefix(domainID, runID, workflowID)
	return fmt.Sprintf("%s_%s.visibility", prefix, t.Format(time.RFC3339))
}

func deserializeQueryVisibilityToken(bytes []byte) (*queryVisibilityToken, error) {
	token := &queryVisibilityToken{}
	err := json.Unmarshal(bytes, token)
	return token, err
}

type parsedVisFilename struct {
	name       string
	closeTime  int64
	LastRunID  string
	WorkflowID string
}

// sortAndFilterFiles sort visibility record file names based on close timestamp (desc) and use hashed runID to break ties.
// if a nextPageToken is give, it only returns filenames that have a smaller close timestamp
func sortAndFilterFiles(filenames []string, token *queryVisibilityToken) ([]string, error) {
	var parsedFilenames []*parsedVisFilename
	for _, name := range filenames {
		pieces := strings.FieldsFunc(name, func(r rune) bool {
			return r == '_' || r == '.'
		})
		if len(pieces) != 5 {
			return nil, fmt.Errorf("failed to parse visibility filename %s", name)
		}

		closeTime, err := time.Parse(time.RFC3339, pieces[3])
		//closeTime, err := strconv.ParseInt(pieces[4], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse visibility filename %s", name)
		}
		parsedFilenames = append(parsedFilenames, &parsedVisFilename{
			name:       name,
			closeTime:  closeTime.UnixNano(),
			WorkflowID: filepath.Base(pieces[1]),
			LastRunID:  pieces[2],
		})
	}

	sort.Slice(parsedFilenames, func(i, j int) bool {
		if parsedFilenames[i].closeTime == parsedFilenames[j].closeTime {
			return parsedFilenames[i].LastRunID > parsedFilenames[j].LastRunID
		}
		return parsedFilenames[i].closeTime > parsedFilenames[j].closeTime
	})

	startIdx := 0
	if token != nil {
		//LastHashedRunID := hash(token.LastRunID)
		startIdx = sort.Search(len(parsedFilenames), func(i int) bool {
			if parsedFilenames[i].closeTime == token.LastCloseTime {
				return parsedFilenames[i].LastRunID < token.LastRunID
			}
			return parsedFilenames[i].closeTime < token.LastCloseTime
		})
	}

	if startIdx == len(parsedFilenames) {
		return []string{}, nil
	}

	var filteredFilenames []string
	for _, parsedFilename := range parsedFilenames[startIdx:] {
		filteredFilenames = append(filteredFilenames, parsedFilename.name)
	}
	return filteredFilenames, nil
}

func matchQuery(record *visibilityRecord, query *parsedQuery) bool {
	if record.CloseTimestamp > query.closeTime || record.CloseTimestamp < query.startTime {
		return false
	}
	if query.workflowID != nil && record.WorkflowID != *query.workflowID {
		return false
	}
	if *query.runID != "" && record.RunID != *query.runID {
		return false
	}
	if query.workflowTypeName != nil && record.WorkflowTypeName != *query.workflowTypeName {
		return false
	}
	if query.closeStatus != nil && record.CloseStatus != *query.closeStatus {
		return false
	}
	return true
}

func convertToExecutionInfo(record *visibilityRecord) *shared.WorkflowExecutionInfo {
	return &shared.WorkflowExecutionInfo{
		Execution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(record.WorkflowID),
			RunId:      common.StringPtr(record.RunID),
		},
		Type: &shared.WorkflowType{
			Name: common.StringPtr(record.WorkflowTypeName),
		},
		StartTime:     common.Int64Ptr(record.StartTimestamp),
		ExecutionTime: common.Int64Ptr(record.ExecutionTimestamp),
		CloseTime:     common.Int64Ptr(record.CloseTimestamp),
		CloseStatus:   record.CloseStatus.Ptr(),
		HistoryLength: common.Int64Ptr(record.HistoryLength),
		Memo:          record.Memo,
		SearchAttributes: &shared.SearchAttributes{
			IndexedFields: archiver.ConvertSearchAttrToBytes(record.SearchAttributes),
		},
	}
}
