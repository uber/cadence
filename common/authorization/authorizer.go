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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination authority_mock.go -self_package github.com/uber/cadence/common/authorization

package authorization

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/uber/cadence/common"

	clientworker "go.uber.org/cadence/worker"

	"github.com/uber/cadence/common/types"
)

const (
	// DecisionDeny means auth decision is deny
	DecisionDeny Decision = iota + 1
	// DecisionAllow means auth decision is allow
	DecisionAllow
)

const (
	// PermissionRead means the user can write on the domain level APIs
	PermissionRead Permission = iota + 1
	// PermissionWrite means the user can write on the domain level APIs
	PermissionWrite
	// PermissionAdmin means the user can read+write on the domain level APIs
	PermissionAdmin
)

type (
	// Attributes is input for authority to make decision.
	// It can be extended in future if required auth on resources like WorkflowType and TaskList
	Attributes struct {
		Actor        string
		APIName      string
		DomainName   string
		WorkflowType *types.WorkflowType
		TaskList     *types.TaskList
		Permission   Permission
		RequestBody  FilteredRequestBody // request object except for data inputs (PII)
	}

	// Result is result from authority.
	Result struct {
		Decision Decision
	}

	// Decision is enum type for auth decision
	Decision int

	// Permission is enum type for auth permission
	Permission int
)

func NewPermission(permission string) Permission {
	switch permission {
	case "read":
		return PermissionRead
	case "write":
		return PermissionWrite
	case "admin":
		return PermissionAdmin
	default:
		return -1
	}
}

// Authorizer is an interface for authorization
type Authorizer interface {
	Authorize(ctx context.Context, attributes *Attributes) (Result, error)
}

func GetAuthProviderClient(privateKey string) (clientworker.AuthorizationProvider, error) {
	pk, err := os.ReadFile(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key path %s", privateKey)
	}
	return clientworker.NewAdminJwtAuthorizationProvider(pk), nil
}

// FilteredRequestBody request object except for data inputs (PII)
type FilteredRequestBody interface {
	SerializeForLogging() (string, error)
}

func validatePermission(claims *JWTClaims, attributes *Attributes, data map[string]string) error {
	groups := ""
	switch attributes.Permission {
	case PermissionRead:
		groups = data[common.DomainDataKeyForReadGroups] + groupSeparator + data[common.DomainDataKeyForWriteGroups]
	case PermissionWrite:
		groups = data[common.DomainDataKeyForWriteGroups]
	default:
		return fmt.Errorf("token doesn't have permission for %v API", attributes.Permission)
	}
	// groups are separated by space
	allowedGroups := strings.Split(groups, groupSeparator)    // groups that allowed by domain configuration(in domainData)
	jwtGroups := strings.Split(claims.Groups, groupSeparator) // groups that the request has associated with

	for _, group1 := range allowedGroups {
		for _, group2 := range jwtGroups {
			if group1 == group2 {
				return nil
			}
		}
	}

	return fmt.Errorf("token doesn't have the right permission, jwt groups: %v, allowed groups: %v", jwtGroups, allowedGroups)
}
