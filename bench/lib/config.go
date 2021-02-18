// Copyright (c) 2017-2021 Uber Technologies Inc.

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

package lib

import (
	"errors"

	"github.com/uber/cadence/common/service/config"
)

type (
	// Config contains the configuration for cadence bench
	Config struct {
		Bench   Bench          `yaml:"bench"`
		Cadence Cadence        `yaml:"cadence"`
		Log     config.Logger  `yaml:"log"`
		Metrics config.Metrics `yaml:"metrics"`
	}

	// Cadence contains the configuration for cadence service
	Cadence struct {
		ServiceName     string `yaml:"service"`
		HostNameAndPort string `yaml:"host"`
	}

	// Bench contains the configuration for bench tests
	Bench struct {
		Name         string   `yaml:"name"`
		Domains      []string `yaml:"domains"`
		NumTaskLists int      `yaml:"numTaskLists"`
	}

	// BasicTestConfig contains the configuration for running the Basic test scenario
	// TODO: update comment
	BasicTestConfig struct {
		TotalLaunchCount                      int     `yaml:"totalLaunchCount"`
		RoutineCount                          int     `yaml:"routineCount"`
		ChainSequence                         int     `yaml:"chainSequence"`
		ConcurrentCount                       int     `yaml:"concurrentCount"`
		PayloadSizeBytes                      int     `yaml:"payloadSizeBytes"`
		MinCadenceSleepInSeconds              int     `yaml:"minCadenceSleepInSeconds"`
		MaxCadenceSleepInSeconds              int     `yaml:"maxCadenceSleepInSeconds"`
		ExecutionStartToCloseTimeoutInSeconds int     `yaml:"executionStartToCloseTimeoutInSeconds"` // default 5m
		ContextTimeoutInSeconds               int     `yaml:"contextTimeoutInSeconds"`               // default 3s
		PanicStressWorkflow                   bool    `yaml:"panicStressWorkflow"`                   // default false
		FailureThreshold                      float64 `yaml:"failureThreshold"`
	}
)

func (c *Config) Validate() error {
	if len(c.Bench.Name) == 0 {
		return errors.New("missing value for bench service name")
	}
	if len(c.Bench.Domains) == 0 {
		return errors.New("missing value for domains property")
	}
	if c.Bench.NumTaskLists == 0 {
		return errors.New("number of taskLists can not be 0")
	}
	return nil
}
