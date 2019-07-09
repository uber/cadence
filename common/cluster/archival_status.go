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

package cluster

import (
	"fmt"
	"strings"
)

type (
	// ArchivalStatus represents the archival status of the cluster
	ArchivalStatus int

	// ArchivalConfig is an immutable representation of the current archival configuration of the cluster
	ArchivalConfig struct {
		status                 ArchivalStatus
		enableReadFromArchival bool
		domainDefaultStatus    ArchivalStatus
		domainDefaultURI       string
	}
)

const (
	// ArchivalDisabled means this cluster is not configured to handle archival
	ArchivalDisabled ArchivalStatus = iota
	// ArchivalPaused means this cluster is configured to handle archival but is currently not archiving
	ArchivalPaused
	// ArchivalEnabled means this cluster is currently archiving
	ArchivalEnabled
)

// NewArchivalConfig constructs a new valid ArchivalConfig
func NewArchivalConfig(
	status ArchivalStatus,
	enableReadFromArchival bool,
	domainDefaultStatus ArchivalStatus,
	domainDefaultURI string,
) *ArchivalConfig {
	ac := &ArchivalConfig{
		status:                 status,
		enableReadFromArchival: enableReadFromArchival,
		domainDefaultStatus:    domainDefaultStatus,
		domainDefaultURI:       domainDefaultURI,
	}
	if !ac.isDomainDefaultValid() {
		ac.domainDefaultStatus = ArchivalDisabled
		ac.domainDefaultURI = ""
	}
	return ac
}

// GetArchivalStatus returns the archival status for ArchivalConfig
func (a *ArchivalConfig) GetArchivalStatus() ArchivalStatus {
	return a.status
}

// ConfiguredForArchival returns true if cluster is configured to handle archival, false otherwise.
func (a *ArchivalConfig) ConfiguredForArchival() bool {
	return a.status != ArchivalDisabled
}

// EnableReadFromArchival indicates whether history can be read from archival
func (a *ArchivalConfig) EnableReadFromArchival() bool {
	return a.enableReadFromArchival
}

func (a *ArchivalConfig) isDomainDefaultValid() bool {
	if a.domainDefaultStatus == ArchivalPaused {
		return false
	}

	URISet := len(a.domainDefaultURI) != 0
	disabled := a.domainDefaultStatus == ArchivalDisabled
	return (!URISet && disabled) || (URISet && !disabled)
}

func getArchivalStatus(str string) (ArchivalStatus, error) {
	str = strings.TrimSpace(strings.ToLower(str))
	switch str {
	case "", "disabled":
		return ArchivalDisabled, nil
	case "paused":
		return ArchivalPaused, nil
	case "enabled":
		return ArchivalEnabled, nil
	}
	return ArchivalDisabled, fmt.Errorf("invalid archival status of %v, valid status are: {\"\", \"disabled\", \"paused\", \"enabled\"}", str)
}
