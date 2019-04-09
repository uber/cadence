package host

import (
	"time"

	"github.com/uber/cadence/common/service/dynamicconfig"
)

var (
	// Override value for integer keys for dynamic config
	intKeys = map[dynamicconfig.Key]int{
		dynamicconfig.FrontendRPS: 1500,
	}
)

type dynamicClient struct {
	client dynamicconfig.Client
}

func (d *dynamicClient) GetValue(name dynamicconfig.Key, defaultValue interface{}) (interface{}, error) {
	return d.client.GetValue(name, defaultValue)
}

func (d *dynamicClient) GetValueWithFilters(
	name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue interface{},
) (interface{}, error) {
	return d.client.GetValueWithFilters(name, filters, defaultValue)
}

func (d *dynamicClient) GetIntValue(name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue int) (int, error) {
	if val, ok := intKeys[name]; ok {
		return val, nil
	}
	return d.client.GetIntValue(name, filters, defaultValue)
}

func (d *dynamicClient) GetFloatValue(name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue float64) (float64, error) {
	return d.client.GetFloatValue(name, filters, defaultValue)
}

func (d *dynamicClient) GetBoolValue(name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue bool) (bool, error) {
	return d.client.GetBoolValue(name, filters, defaultValue)
}

func (d *dynamicClient) GetStringValue(name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue string) (string, error) {
	return d.client.GetStringValue(name, filters, defaultValue)
}

func (d *dynamicClient) GetMapValue(
	name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	return d.client.GetMapValue(name, filters, defaultValue)
}

func (d *dynamicClient) GetDurationValue(
	name dynamicconfig.Key, filters map[dynamicconfig.Filter]interface{}, defaultValue time.Duration,
) (time.Duration, error) {
	return d.client.GetDurationValue(name, filters, defaultValue)
}

// newIntegrationConfigClient - returns a dynamic config client for integration testing
func newIntegrationConfigClient(client dynamicconfig.Client) dynamicconfig.Client {
	return &dynamicClient{client}
}
