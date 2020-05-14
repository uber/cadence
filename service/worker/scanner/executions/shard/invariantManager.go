package shard

import (
	"fmt"

	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"github.com/uber/cadence/service/worker/scanner/executions/invariants"
)

type (
	invariantManager struct {
		invariants []common.Invariant
		types []common.InvariantType
	}
)

// NewInvariantManager handles running a collection of invariants according to the policy provided.
// InvariantManager takes care of ensuring invariants are run the their correct dependency order.
func NewInvariantManager(
	policy common.InvariantPolicy,
	pr common.PersistenceRetryer,
) common.InvariantManager {
	manager := &invariantManager{}
	manager.invariants, manager.types = getSortedInvariants(policy, pr)
	return manager
}


func (i *invariantManager) RunChecks(execution common.Execution) common.CheckResult {
	resources := &common.InvariantResourceBag{}
	details := make(map[common.InvariantType]string)
	for _, iv := range i.invariants {
		result := iv.Check(execution, resources)
		if result.CheckResultType != common.CheckResultTypeHealthy {
			return result
		}
		details[iv.InvariantType()] = result.Info
	}
	return common.CheckResult{
		CheckResultType: common.CheckResultTypeHealthy,
		Info: fmt.Sprintf("all invariants ran were healthy: %v", i.types),
		InfoDetails: fmt.Sprintf("%v", details),
	}
}

func (i *invariantManager) RunFixes(execution common.Execution) common.FixResult {
	for _, iv := range i.invariants {
		result := iv.Fix(execution)

	}



	return common.FixResult{}
}

func (i *invariantManager) InvariantTypes() []common.InvariantType {
	return i.types
}

func getSortedInvariants(policy common.InvariantPolicy, pr common.PersistenceRetryer) ([]common.Invariant, []common.InvariantType) {
	var ivs []common.Invariant
	switch policy {
	case common.InvariantPolicyAll:
		ivs = []common.Invariant{invariants.NewHistoryExists(pr), invariants.NewValidFirstEvent(pr), invariants.NewOpenCurrentExecution(pr)}
	case common.InvariantPolicySkipHistory:
		ivs = []common.Invariant{invariants.NewOpenCurrentExecution(pr)}
	default:
		panic("unknown policy type")
	}
	types := make([]common.InvariantType, len(ivs), len(ivs))
	for i, iv := range ivs {
		types[i] = iv.InvariantType()
	}
	return ivs, types
}