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

	"github.com/uber/cadence/common/config"
)

const (
	// EnvKeyRoot the environment variable key for runtime root dir
	EnvKeyRoot = "CADENCE_BENCH_ROOT"
	// EnvKeyConfigDir the environment variable key for config dir
	EnvKeyConfigDir = "CADENCE_BENCH_CONFIG_DIR"
	// EnvKeyEnvironment is the environment variable key for environment
	EnvKeyEnvironment = "CADENCE_BENCH_ENVIRONMENT"
	// EnvKeyAvailabilityZone is the environment variable key for AZ
	EnvKeyAvailabilityZone = "CADENCE_BENCH_AVAILABILITY_ZONE"
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

	// TODO: add comment for each config field

	// CronTestConfig contains the configuration for running a set of
	// testsuites in parallel based on a cron schedule
	CronTestConfig struct {
		TestSuites []TestSuiteConfig `yaml:"testSuites"`
	}

	// TestSuiteConfig contains the configration for running a set of
	// tests sequentially
	TestSuiteConfig struct {
		Name    string                 `yaml:"name"`
		Domain  string                 `yaml:"domain"`
		Configs []AggregatedTestConfig `yaml:"configs"`
	}

	// AggregatedTestConfig contains the configration for a single test
	AggregatedTestConfig struct {
		Name             string                    `yaml:"name"`
		Description      string                    `yaml:"description"`
		TimeoutInSeconds int32                     `yaml:"timeoutInSeconds"`
		Basic            *BasicTestConfig          `yaml:"basic"`
		Signal           *SignalTestConfig         `yaml:"signal"`
		Timer            *TimerTestConfig          `yaml:"timer"`
		ConcurrentExec   *ConcurrentExecTestConfig `yaml:"concurrentExec"`
		Cancellation     *CancellationTestConfig   `yaml:"cancellation"`
	}

	// BasicTestConfig contains the configuration for running the Basic test scenario
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

	// SignalTestConfig is the parameters for signalLoadTestWorkflow
	SignalTestConfig struct {
		// LoaderCount defines how many loader activities
		LoaderCount int `yaml:"loaderCount"`
		// LoadTestWorkflowCount defines how many load test workflow in total
		LoadTestWorkflowCount int `yaml:"loadTestWorkflowCount"`
		// SignalCount is the number of signals per workflow
		SignalCount int `yaml:"signalCount"`
		// SignalDataSize is the size of signal data
		SignalDataSize int `yaml:"signalDataSize"`
		// RateLimit is per loader rate limit to hit cadence server
		RateLimit                         int     `yaml:"rateLimit"`
		WorkflowExecutionTimeoutInSeconds int     `yaml:"workflowExecutionTimeoutInSeconds"`
		DecisionTaskTimeoutInSeconds      int     `yaml:"decisionTaskTimeoutInSeconds"`
		FailureThreshold                  float64 `yaml:"failureThreshold"`
		ProcessSignalWorkflowConfig
	}

	// ProcessSignalWorkflowConfig is the parameters to process signal workflow
	ProcessSignalWorkflowConfig struct {
		// CampaignCount is the number of local activities to be executed
		CampaignCount int `yaml:"campaignCount"`
		// ActionRate is probability that local activities result in actual action
		ActionRate float64 `yaml:"actionRate"`
		// Local activity failure rate
		FailureRate float64 `yaml:"failureRate"`
		// SignalCount before continue as new
		SignalBeforeContinueAsNew int   `yaml:"signalBeforeContinueAsNew"`
		EnableRollingWindow       bool  `yaml:"enableRollingWindow"`
		ScheduleTimeNano          int64 `yaml:"scheduleTimeNano"`
		MaxSignalDelayInSeconds   int   `yaml:"maxSignalDelayInSeconds"`
		MaxSignalDelayCount       int   `yaml:"maxSignalDelayCount"`
	}

	// TimerTestConfig contains the config for running timer bench test
	TimerTestConfig struct {
		// TotalTimerCount is the total number of timers to fire
		TotalTimerCount int `yaml:"totalTimerCount"`

		// TimerPerWorkflow is the number of timers in each workflow
		// workflow will continue execution and complete when the first timer fires
		// Set this number larger than one to test no-op timer case
		// TotalTimerCount / TimerPerWorkflow = total number of workflows
		TimerPerWorkflow int `yaml:"timerPerWorkflow"`

		// ShortestTimerDurationInSeconds after test start, the first timer will fire
		ShortestTimerDurationInSeconds int `yaml:"shortestTimerDurationInSeconds"`

		// LongestTimerDurationInSeconds after test start, the last timer will fire
		LongestTimerDurationInSeconds int `yaml:"longestTimerDurationInSeconds"`

		// MaxTimerLatencyInSeconds specifies the maximum latency for the first timer in the workflow
		// if a timer's latency is larger than this value, that timer will be considered as failed
		// if > 1% timer fire beyond this threshold, the test will fail
		MaxTimerLatencyInSeconds int `yaml:"maxTimerLatencyInSeconds"`

		// TimerTimeoutInSeconds specifies the duration beyond which a timer is considered as lost and fail the test
		TimerTimeoutInSeconds int `yaml:"timerTimeoutInSeconds"`

		// RoutineCount is the number of goroutines used for starting workflows
		// approx. RPS = 10 * RoutineCount
		// # of workflows = TotalTimerCount / TimerPerWorkflow
		// please make sure ShortestTimerDurationInSeconds > # of workflows / RPS, so that timer starts firing
		// after all workflows has been started.
		// please also make sure test timeout > LongestTimerDurationInSeconds + TimerTimeoutInSeconds
		RoutineCount int `yaml:"routineCount"`
	}

	// ConcurrentExecTestConfig contains the config for running concurrent execution test
	ConcurrentExecTestConfig struct {
		// TotalBatches is the total number of batches
		TotalBatches int `yaml:"totalBatches"`

		// Concurrency specifies the number of batches that will be run concurrently
		Concurrency int `yaml:"concurrency"`

		// BatchType specifies the type of batch, can be either "activity" or "childWorkflow", case insensitive
		BatchType string `yaml:"batchType"`

		// BatchSize specifies the number of activities or childWorkflows scheduled in a single decision batch
		BatchSize int `yaml:"batchSize"`

		// BatchPeriodInSeconds specifies the time interval between two set of batches (each set has #concurrency batches)
		BatchPeriodInSeconds int `yaml:"batchPeriodInSeconds"`

		// BatchMaxLatencyInSeconds specifies the max latency for scheduling/starting the activity or childWorkflow.
		// if latency is higher than this number, the corresponding activity or childWorkflow will be
		// considered as failed. This bench test is considered as success if:
		// avg(succeed activity or childWorkflow / batchSize) >= 0.99
		// If any of the activity or childWorkflow returns an error (for example, execution takes longer than BatchTimeoutInSeconds),
		// the bench test will fail immediately
		BatchMaxLatencyInSeconds int `yaml:"batchMaxLatencyInSeconds"`

		// BatchTimeoutInSeconds specifies the timeout for each batch execution
		BatchTimeoutInSeconds int `yaml:"batchTimeoutInSeconds"`
	}

	// CancellationTestConfig contains the config for running workflow cancellation test
	CancellationTestConfig struct {
		// TotalLaunchCount is the total number of workflow to start
		// note: make sure TotalLaunchCount mod Concurrency = 0 otherwise the validation for the test will fail
		TotalLaunchCount int `yaml:"totalLaunchCount"`

		// Concurrency specifies the concurrency for start and cancel workflow execution
		// approx. RPS = 10 * concurrency, approx. duration = totalLaunchCount / RPS
		Concurrency int `yaml:"concurrency"`

		// ContextTimeoutInSeconds specifies the context timeout for start and cancel workflow execution call
		// default: 3s
		ContextTimeoutInSeconds int `yaml:"contextTimeoutInSeconds"`
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
