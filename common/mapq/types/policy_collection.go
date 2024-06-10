// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

import (
	"fmt"
	"sort"
	"strings"
)

type NodePolicyCollection struct {
	// policies is sorted by level and generic to special
	policies []NodePolicy
}

func NewNodePolicyCollection(policies []NodePolicy) NodePolicyCollection {
	trimPaths(policies)
	sortPolicies(policies)
	return NodePolicyCollection{
		policies: policies,
	}
}

func (npc NodePolicyCollection) GetMergedPolicyForNode(path string) (NodePolicy, error) {
	result := NodePolicy{}
	for _, policy := range npc.policies {
		pathParts := strings.Split(path, "/")
		plcPathParts := strings.Split(policy.Path, "/")
		length := len(pathParts)
		if len(plcPathParts) <= length {
			length = len(plcPathParts)
		} else {
			// if policy path is longer than the node path then it can't be a match
			continue
		}

		match := true
		for i := 0; i < length; i++ {
			pathPart := pathParts[i]
			plcPathPart := plcPathParts[i]
			if pathPart == "." || plcPathPart == "." || pathPart == plcPathPart {
				continue
			}
			match = false
			break
		}

		if !match {
			continue
		}
		var err error
		result, err = result.Merge(policy)
		if err != nil {
			return result, fmt.Errorf("failed to merge policies: %w", err)
		}
	}

	// set the path in the result
	result.Path = path
	return result, nil
}

func (npc NodePolicyCollection) GetPolicies() []NodePolicy {
	// return a copy of the slice so the order is not messed up
	policies := make([]NodePolicy, len(npc.policies))
	copy(policies, npc.policies)
	return policies
}

func trimPaths(policies []NodePolicy) {
	for i := range policies {
		policies[i].Path = strings.TrimRight(policies[i].Path, "/")
	}
}

func sortPolicies(policies []NodePolicy) {
	// sort policies by level and generic to special
	sort.Slice(policies, func(i, j int) bool {
		parts1 := strings.Split(policies[i].Path, "/")
		parts2 := strings.Split(policies[j].Path, "/")
		l1 := len(parts1)
		l2 := len(parts2)
		if l1 != l2 {
			return l1 < l2
		}

		pathPartPrecedence := map[any]int{
			".": 0,
			"*": 1,
		}
		for partIdx := 0; partIdx < l1; partIdx++ {
			p1, ok := pathPartPrecedence[parts1[partIdx]]
			if !ok {
				p1 = 2
			}
			p2, ok := pathPartPrecedence[parts2[partIdx]]
			if !ok {
				p2 = 2
			}

			if p1 != p2 {
				return p1 < p2
			}
		}

		// If parts have same precedence then just return the order
		return i < j
	})
}
