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

package config

import (
	"errors"
)

const (
	// ArchivalEnabled is the status for enabling archival
	ArchivalEnabled = "enabled"
)

// ValidateProvider validates the provider section of the archival config
func (a *Archival) ValidateProvider() error {
	historyArchivalEnabled := a.History.Status == ArchivalEnabled
	if (historyArchivalEnabled && a.History.Provider == nil) || (!historyArchivalEnabled && a.History.Provider != nil) {
		return errors.New("Invalid history archival config")
	}

	visibilityArchivalEnabled := a.Visibility.Status == ArchivalEnabled
	if (visibilityArchivalEnabled && a.Visibility.Provider == nil) || (!visibilityArchivalEnabled && a.Visibility.Provider != nil) {
		return errors.New("Invalid visibility archival config")
	}

	return nil
}
