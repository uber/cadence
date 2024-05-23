package mapq

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/playground/mapq/types"
)

func TestExample(t *testing.T) {
	cl, err := New(
		WithConsumerFactory(&NoOpConsumerFactory{}),
		WithPersister(&InMemoryPersister{}),
		WithPartitions([]string{"type", "sub-type", "domain"}),
		WithPolicies([]types.NodePolicy{
			// level 0: split by type
			{
				Path: "",
				SplitPolicy: &types.SplitPolicy{
					PredefinedSplits: []any{"timer", "transfer"},
				},
			},
			// level 1: timer node
			{
				Path: "/timer",
				SplitPolicy: &types.SplitPolicy{
					PredefinedSplits: []any{
						persistence.TaskTypeDeleteHistoryEvent,
					},
				},
			},
			// level 1: transfer node
			{
				Path: "/transfer",
				SplitPolicy: &types.SplitPolicy{
					PredefinedSplits: []any{
						persistence.TransferTaskTypeStartChildExecution,
					},
				},
			},
			// level 2: nodes per <type, sub-type> pairs
			// - default 1000 RPS for per sub-type node
			// - split by domain
			// - allow top 5 domains to be split based on burst
			{
				Path: "/./.",
				SplitPolicy: &types.SplitPolicy{
					MaxNumChildren: 5, // allow top 4 tomains to be split. 5th domain will be catch-all
					Strategy: &types.SplitStrategy{
						SplitEnqueueRPSThreshold: 100,
					},
				},
			},
			// override for timer delete history event:
			// - only allow 50 RPS
			// - disable split policy
			{
				Path:           "/timer/4",
				DispatchPolicy: &types.DispatchPolicy{DispatchRPS: 50},
				SplitPolicy:    &types.SplitPolicy{Disabled: true},
			},
			// override for start child execution
			// - only allow 10 RPS
			// - disable split policy
			{
				Path:           "/transfer/4",
				DispatchPolicy: &types.DispatchPolicy{DispatchRPS: 10},
				SplitPolicy:    &types.SplitPolicy{Disabled: true},
			},
			// level 3: default policy for all nodes at this level. nodes per <type, sub-type, domain> pairs
			// - only allow 100 rps
			// - disable split policy
			{
				Path:           "/././.",
				DispatchPolicy: &types.DispatchPolicy{DispatchRPS: 100},
				SplitPolicy:    &types.SplitPolicy{Disabled: true},
			},
			// level 3: override policy for catch-all nodes at this level (all domains that don't have a specific node)
			// this policy will override the 100 RPS policy defined above to give more RPS to catch-all nodes
			{
				Path:           "/././*",
				DispatchPolicy: &types.DispatchPolicy{DispatchRPS: 1000},
				SplitPolicy:    &types.SplitPolicy{Disabled: true},
			},
		}),
	)

	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err := cl.Start(ctx); err != nil {
		panic(err)
	}
	defer cl.Stop(ctx)

	err = cl.Enqueue(context.Background(), []types.Item{
		newTimerItem("d1", time.Now(), persistence.TaskTypeDecisionTimeout),
		newTimerItem("d1", time.Now(), persistence.TaskTypeActivityTimeout),
		newTimerItem("d1", time.Now(), persistence.TaskTypeUserTimer),
	})
	if err != nil {
		panic(err)
	}

	err = cl.Enqueue(context.Background(), []types.Item{
		newTransferItem("d2", 1, persistence.TransferTaskTypeDecisionTask),
		newTransferItem("d2", 2, persistence.TransferTaskTypeActivityTask),
		newTransferItem("d2", 3, persistence.TransferTaskTypeStartChildExecution),
		newTransferItem("d2", 4, persistence.TransferTaskTypeStartChildExecution),
		newTransferItem("d2", 5, persistence.TransferTaskTypeStartChildExecution),
	})
	if err != nil {
		panic(err)
	}

}

var _ types.ConsumerFactory = (*NoOpConsumerFactory)(nil)

type NoOpConsumerFactory struct{}

func (f *NoOpConsumerFactory) New(context.Context, types.ItemPartitions) types.Consumer {
	return &NoOpConsumer{}
}

type NoOpConsumer struct{}

func (c *NoOpConsumer) Start(context.Context) error {
	return nil
}

func (c *NoOpConsumer) Stop(context.Context) error {
	return nil
}

func (c *NoOpConsumer) Process(ctx context.Context, item types.Item) error {
	fmt.Printf("processing item: %v\n", item)
	return nil
}

type InMemoryPersister struct {
	items   []types.ItemToPersist
	offsets *types.Offsets
}

func (p *InMemoryPersister) Persist(ctx context.Context, items []types.ItemToPersist) error {
	fmt.Printf("persisting %v items\n", len(items))
	return nil
}

func (p *InMemoryPersister) GetOffsets(context.Context) (*types.Offsets, error) {
	return p.offsets, nil
}

func (p *InMemoryPersister) CommitOffsets(ctx context.Context, offsets *types.Offsets) error {
	fmt.Printf("committing offsets: %v\n", offsets)
	p.offsets = offsets
	return nil
}

func newTimerItem(domain string, t time.Time, timerType int) types.Item {
	switch timerType {
	case persistence.TaskTypeDecisionTimeout:
	case persistence.TaskTypeActivityTimeout:
	case persistence.TaskTypeUserTimer:
	case persistence.TaskTypeWorkflowTimeout:
	case persistence.TaskTypeDeleteHistoryEvent:
	case persistence.TaskTypeActivityRetryTimer:
	case persistence.TaskTypeWorkflowBackoffTimer:
	default:
		panic(fmt.Errorf("unknown timer type: %v", timerType))
	}

	return &timerItem{
		t:         t,
		timerType: timerType,
		domain:    domain,
	}
}

type timerItem struct {
	t         time.Time
	timerType int
	domain    string
}

func (t *timerItem) String() string {
	return fmt.Sprintf("timerItem{timerType: %v, time: %v}", t.timerType, t.t)
}

func (t *timerItem) Offset() int64 {
	return t.t.UnixNano()
}

func (t *timerItem) GetAttribute(key string) any {
	switch key {
	case "type":
		return "timer"
	case "sub-type":
		return t.timerType
	case "domain":
		return t.domain
	default:
		panic(fmt.Errorf("unknown key: %v", key))
	}
}

func newTransferItem(domain string, taskID int64, transferType int) types.Item {
	switch transferType {
	case persistence.TransferTaskTypeActivityTask:
	case persistence.TransferTaskTypeDecisionTask:
	case persistence.TransferTaskTypeCloseExecution:
	case persistence.TransferTaskTypeCancelExecution:
	case persistence.TransferTaskTypeSignalExecution:
	case persistence.TransferTaskTypeStartChildExecution:
	case persistence.TransferTaskTypeRecordWorkflowStarted:
	case persistence.TransferTaskTypeResetWorkflow:
	case persistence.TransferTaskTypeUpsertWorkflowSearchAttributes:
	case persistence.TransferTaskTypeRecordWorkflowClosed:
	case persistence.TransferTaskTypeRecordChildExecutionCompleted:
	case persistence.TransferTaskTypeApplyParentClosePolicy:
	default:
		panic(fmt.Errorf("unknown transfer type: %v", transferType))
	}

	return &transferItem{
		domain:       domain,
		taskID:       taskID,
		transferType: transferType,
	}
}

type transferItem struct {
	taskID       int64
	transferType int
	domain       string
}

func (t *transferItem) String() string {
	return fmt.Sprintf("transferItem{transferType: %v, taskID: %v}", t.transferType, t.taskID)
}

func (t *transferItem) Offset() int64 {
	return t.taskID
}

func (t *transferItem) GetAttribute(key string) any {
	switch key {
	case "type":
		return "transfer"
	case "sub-type":
		return t.transferType
	case "domain":
		return t.domain
	default:
		panic(fmt.Errorf("unknown key: %v", key))
	}
}
