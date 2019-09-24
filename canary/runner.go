package canary

import (
	"sync"

	"go.uber.org/zap"
)

type canaryRunner struct {
	*RuntimeContext
	config *Config
}

// New creates and returns a runnable which spins
// up a set of canaries based on supplied config
func newCanaryRunner(cfg *Config, runtime *RuntimeContext) Runnable {
	return &canaryRunner{
		RuntimeContext: runtime,
		config:         cfg,
	}
}

// Run runs the canaries
func (r *canaryRunner) Run() error {
	r.metrics.Counter("restarts").Inc(1)
	if len(r.config.Excludes) != 0 {
		updateSanityChildWFList(r.config.Excludes)
	}

	var wg sync.WaitGroup
	for _, d := range r.config.Domains {
		canary := newCanary(d, r.RuntimeContext)
		r.logger.Info("starting canary", zap.String("domain", d))
		r.execute(canary, &wg)
	}
	wg.Wait()
	return nil
}

func (r *canaryRunner) execute(task Runnable, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		task.Run()
		wg.Done()
	}()
}

func updateSanityChildWFList(excludes []string) {
	var temp []string
	for _, childName := range sanityChildWFList {
		if !isStringInList(childName, excludes) {
			temp = append(temp, childName)
		}
	}
	sanityChildWFList = temp
}

func isStringInList(str string, list []string) bool {
	for _, l := range list {
		if l == str {
			return true
		}
	}
	return false
}
