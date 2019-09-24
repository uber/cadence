package canary

import (
	"fmt"
	"time"

	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/config"
	"go.uber.org/zap"
)

type (
	// Config contains the configurable yaml
	// properties for the canary runtime
	Config struct {
		Domains  []string `yaml:"domains"`
		Excludes []string `yaml:"excludes"`
	}
)

const (
	// ConfigurationKey is the config YAML key for the canary module
	ConfigurationKey = "canary"
)

// Init validates and initializes the config
func newCanaryConfig(provider config.Provider) (*Config, error) {
	raw := provider.Get(ConfigurationKey)
	var cfg Config
	if err := raw.Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load canary configuration with error: %v", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Domains) == 0 {
		return fmt.Errorf("missing value for domains property")
	}
	return nil
}

// RuntimeContext contains all the context
// information needed to run the canary
type RuntimeContext struct {
	Env     string
	logger  *zap.Logger
	metrics tally.Scope
	service workflowserviceclient.Interface
}

// NewRuntimeContext builds a runtime context from the config
func newRuntimeContext(env string, logger *zap.Logger, scope tally.Scope, service workflowserviceclient.Interface) *RuntimeContext {
	return &RuntimeContext{
		Env:     env,
		logger:  logger,
		metrics: scope,
		service: service,
	}
}

// newTallyScope builds and returns a tally scope from m3 configuration
func newTallyScope(cfg *m3.Configuration) (tally.Scope, error) {
	reporter, err := cfg.NewReporter()
	if err != nil {
		return nil, err
	}
	scopeOpts := tally.ScopeOptions{CachedReporter: reporter}
	scope, _ := tally.NewRootScope(scopeOpts, time.Second)
	return scope, nil
}
