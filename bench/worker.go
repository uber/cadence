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

package bench

import (
	"context"
	"sync"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/uber/cadence/bench/lib"
	"github.com/uber/cadence/bench/load/basic"
	"github.com/uber/cadence/bench/load/cancellation"
	"github.com/uber/cadence/bench/load/common"
	"github.com/uber/cadence/bench/load/concurrentexec"
	"github.com/uber/cadence/bench/load/cron"
	"github.com/uber/cadence/bench/load/signal"
	"github.com/uber/cadence/bench/load/timer"
)

type (
	loadTestWorker struct {
		clients map[string]lib.CadenceClient
		runtime *lib.RuntimeContext

		shutdownWG sync.WaitGroup
	}
)

const (
	decisionWorkerConcurrency = 20
	activityWorkerConcurrency = 5000
	stickyWorkflowCacheSize   = 100000
)

// NewWorker builds and returns a new instance of load test worker
func NewWorker(cfg *lib.Config) (lib.Runnable, error) {
	rc, err := lib.NewRuntimeContext(cfg)
	if err != nil {
		return nil, err
	}

	clients, err := lib.NewCadenceClients(rc)
	if err != nil {
		return nil, err
	}

	return &loadTestWorker{
		runtime: rc,
		clients: clients,
	}, nil
}

// Run runs the worker
func (w *loadTestWorker) Run() error {
	if err := w.createDomains(); err != nil {
		w.runtime.Logger.Error("Failed to create bench domain", zap.Error(err))
	}

	worker.SetStickyWorkflowCacheSize(stickyWorkflowCacheSize)
	w.runtime.Metrics.Counter("worker.restarts").Inc(1)
	for _, domainName := range w.runtime.Bench.Domains {
		for i := 0; i < w.runtime.Bench.NumTaskLists; i++ {
			w.shutdownWG.Add(1)
			go w.runWorker(domainName, common.GetTaskListName(i))
		}
	}

	w.shutdownWG.Wait()
	return nil
}

func (w *loadTestWorker) createDomains() error {
	for _, domainName := range w.runtime.Bench.Domains {
		desc := "Domain for running cadence load test"
		owner := "cadence-bench"
		if err := w.clients[domainName].CreateDomain(domainName, desc, owner); err != nil {
			return err
		}
	}

	return nil
}

func (w *loadTestWorker) runWorker(
	domainName string,
	taskList string,
) {
	defer w.shutdownWG.Done()

	opts := w.newWorkerOptions(domainName)
	worker := worker.New(w.clients[domainName].Service, domainName, taskList, opts)
	registerWorkers(worker)
	registerLaunchers(worker)

	if err := worker.Run(); err != nil {
		w.runtime.Logger.Error("Failed to start worker", zap.String("Domain", domainName), zap.String("TaskList", taskList), zap.Error(err))
	}
}

func (w *loadTestWorker) newWorkerOptions(domainName string) worker.Options {
	return worker.Options{
		Logger:                    w.runtime.Logger,
		MetricsScope:              w.runtime.Metrics,
		BackgroundActivityContext: w.newActivityContext(domainName),
		// TODO: do we need these limits?
		MaxConcurrentActivityExecutionSize:     activityWorkerConcurrency,
		MaxConcurrentDecisionTaskExecutionSize: decisionWorkerConcurrency,
	}
}

func (w *loadTestWorker) newActivityContext(domainName string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, lib.CtxKeyCadenceClient, w.clients[domainName])
	return context.WithValue(ctx, lib.CtxKeyRuntimeContext, w.runtime)
}

func registerWorkers(w worker.Worker) {
	common.RegisterWorker(w)
	basic.RegisterWorker(w)
	signal.RegisterWorker(w)
	timer.RegisterWorker(w)
	concurrentexec.RegisterWorker(w)
	cancellation.RegisterWorker(w)
}

func registerLaunchers(w worker.Worker) {
	cron.RegisterLauncher(w)
	signal.RegisterLauncher(w)
	basic.RegisterLauncher(w)
	timer.RegisterLauncher(w)
	concurrentexec.RegisterLauncher(w)
	cancellation.RegisterLauncher(w)
}
