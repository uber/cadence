// Copyright (c) 2019 Uber Technologies, Inc.
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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination nDCEventsReapplier_mock.go

package history

import (
	ctx "context"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

type (
	nDCEventsReapplier interface {
		reapplyEvents(ctx ctx.Context, msBuilder mutableState, historyEventsBlob *shared.DataBlob) error
	}

	nDCEventsReapplierImpl struct {
		historySerializer persistence.PayloadSerializer
		metricsClient     metrics.Client
		logger            log.Logger
	}
)

func newNDCEventsReapplier(
	metricsClient metrics.Client,
	logger log.Logger,
) nDCEventsReapplier {

	return &nDCEventsReapplierImpl{
		historySerializer: persistence.NewPayloadSerializer(),
		metricsClient:     metricsClient,
		logger:            logger,
	}
}

func (r *nDCEventsReapplierImpl) reapplyEvents(
	ctx ctx.Context,
	msBuilder mutableState,
	historyEventsBlob *shared.DataBlob,
) error {

	reapplyEvents, err := r.validateHistoryEvents(historyEventsBlob)
	if err != nil {
		return err
	}
	if len(reapplyEvents) == 0 {
		return nil
	}

	if !msBuilder.IsWorkflowExecutionRunning() {
		// TODO when https://github.com/uber/cadence/issues/2420 is finished
		//  reset to workflow finish event
		//  ignore this case for now
		return nil
	}

	// TODO: need to have signal deduplicate logic
	for _, event := range reapplyEvents {
		signal := event.GetWorkflowExecutionSignaledEventAttributes()
		if _, err := msBuilder.AddWorkflowExecutionSignaled(
			signal.GetSignalName(),
			signal.GetInput(),
			signal.GetIdentity(),
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *nDCEventsReapplierImpl) validateHistoryEvents(
	historyEventsBlob *shared.DataBlob,
) ([]*shared.HistoryEvent, error) {
	if historyEventsBlob == nil {
		return nil, nil
	}

	historyEvents, err := r.historySerializer.DeserializeBatchEvents(&persistence.DataBlob{
		Encoding: common.EncodingTypeThriftRW,
		Data:     historyEventsBlob.Data,
	})
	if err != nil {
		return nil, err
	}
	reapplyEvents := []*shared.HistoryEvent{}
	// TODO: need to implement Reapply policy
	for _, event := range historyEvents {
		switch event.GetEventType() {
		case shared.EventTypeWorkflowExecutionSignaled:
			reapplyEvents = append(reapplyEvents, event)
		}
	}
	return reapplyEvents, nil
}
