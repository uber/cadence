package invariants

type TimeoutType string

const (
	TimeoutTypeExecution     TimeoutType = "The Workflow Execution has timed out"
	TimeoutTypeActivity      TimeoutType = "Activity task has timed out"
	TimeoutTypeDecision      TimeoutType = "Decision task has timed out"
	TimeoutTypeChildWorkflow TimeoutType = "Child Workflow Execution has timed out"
)

func (tt TimeoutType) String() string {
	return string(tt)
}
