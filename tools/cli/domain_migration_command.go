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

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common/dynamicconfig"
	dc "github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
)

const (
	// workflows running longer than 14 days are considered long running workflows
	longRunningDuration = 14 * 24 * time.Hour
)

var (
	domainMigrationTemplate = `Validation Check:
{{- range .}}
- {{.ValidationCheck}}: {{.ValidationResult}}
{{- with .ValidationDetails}}
  {{- with .CurrentDomainRow}}
    Current Domain:
      Name: {{.DomainInfo.Name}}
      UUID: {{.DomainInfo.UUID}}
  {{- end}}
  {{- with .NewDomainRow}}
    New Domain:
      Name: {{.DomainInfo.Name}}
      UUID: {{.DomainInfo.UUID}}
  {{- end}}
  {{- if ne (len .MismatchedDomainMetaData) 0 }}   
    Mismatched Domain Meta Data: {{.MismatchedDomainMetaData}}
  {{- end }}
  {{- if .LongRunningWorkFlowNum}}
    Long Running Workflow Num (> 14 days): {{.LongRunningWorkFlowNum}}
  {{- end}}
  {{- if .MissingCurrSearchAttributes}}
    Missing Search Attributes in Current Domain:
    {{- range .MissingCurrSearchAttributes}}
      - {{.}}
    {{- end}}
  {{- end}}
  {{- if .MissingNewSearchAttributes}}
    Missing Search Attributes in New Domain:
    {{- range .MissingNewSearchAttributes}}
      - {{.}}
    {{- end}}
  {{- end}}
  {{- range .MismatchedDynamicConfig}}
    {{- $dynamicConfig := . }}
    - Config Key: {{.Key}}
      {{- range $i, $v := .CurrValues}}
        Current Response:
          Data: {{ printf "%s" (index $dynamicConfig.CurrValues $i).Value.Data }}
          Filters:
          {{- range $filter := (index $dynamicConfig.CurrValues $i).Filters}}
            - Name: {{ $filter.Name }}
              Value: {{ printf "%s" $filter.Value.Data }}
          {{- end}}
        New Response:
          Data: {{ printf "%s" (index $dynamicConfig.NewValues $i).Value.Data }}
          Filters:
          {{- range $filter := (index $dynamicConfig.NewValues $i).Filters}}
            - Name: {{ $filter.Name }}
              Value: {{ printf "%s" $filter.Value.Data }}
          {{- end}}
        {{- end}}
      {{- end}}
  {{- end}}
{{- end}}
`
	emptyGetDynamicConfigRequest = &types.GetDynamicConfigResponse{
		Value: &types.DataBlob{
			EncodingType: types.EncodingTypeJSON.Ptr(),
		},
	}
)

type DomainMigrationCommand interface {
	Validation(c *cli.Context)
	DomainMetaDataCheck(c *cli.Context) DomainMigrationRow
	DomainWorkFlowCheck(c *cli.Context) DomainMigrationRow
	SearchAttributesChecker(c *cli.Context) DomainMigrationRow
	DynamicConfigCheck(c *cli.Context) DomainMigrationRow
}

func (d *domainMigrationCLIImpl) NewDomainMigrationCLIImpl(c *cli.Context) *domainMigrationCLIImpl {
	return &domainMigrationCLIImpl{
		frontendClient:         cFactory.ServerFrontendClient(c),
		destinationClient:      cFactory.ServerFrontendClientForMigration(c),
		frontendAdminClient:    cFactory.ServerAdminClient(c),
		destinationAdminClient: cFactory.ServerAdminClientForMigration(c),
	}
}

// Export a function to create an instance of the domainMigrationCLIImpl.
func NewDomainMigrationCommand(c *cli.Context) DomainMigrationCommand {
	return &domainMigrationCLIImpl{}
}

type domainMigrationCLIImpl struct {
	frontendClient, destinationClient           frontend.Client
	frontendAdminClient, destinationAdminClient admin.Client
}

func (d *domainMigrationCLIImpl) Validation(c *cli.Context) {
	checkers := []func(*cli.Context) DomainMigrationRow{
		d.DomainMetaDataCheck,
		d.DomainWorkFlowCheck,
		d.DynamicConfigCheck,
		d.SearchAttributesChecker,
	}
	wg := &sync.WaitGroup{}
	results := make([]DomainMigrationRow, len(checkers))
	for i := range checkers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = checkers[i](c)
		}(i)
	}
	wg.Wait()

	renderOpts := RenderOptions{
		DefaultTemplate: domainMigrationTemplate,
		Color:           true,
		Border:          true,
		PrintDateTime:   true,
	}

	if err := Render(c, results, renderOpts); err != nil {
		ErrorAndExit("Failed to render", err)
	}
}

func (d *domainMigrationCLIImpl) DomainMetaDataCheck(c *cli.Context) DomainMigrationRow {
	domain := c.GlobalString(FlagDomain)
	newDomain := c.String(FlagDestinationDomain)
	ctx, cancel := newContext(c)
	defer cancel()
	currResp, err := d.frontendClient.DescribeDomain(ctx, &types.DescribeDomainRequest{
		Name: &domain,
	})
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Could not describe old domain, Please check to see if old domain exists before migrating."), err)
	}
	newResp, err := d.destinationClient.DescribeDomain(ctx, &types.DescribeDomainRequest{
		Name: &newDomain,
	})
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Could not describe new domain, Please check to see if new domain exists before migrating."), err)
	}
	validationResult, mismatchedMetaData := metaDataValidation(currResp, newResp)
	validationRow := DomainMigrationRow{
		ValidationCheck:  "Domain Meta Data",
		ValidationResult: validationResult,
		ValidationDetails: ValidationDetails{
			CurrentDomainRow:         currResp,
			NewDomainRow:             newResp,
			MismatchedDomainMetaData: mismatchedMetaData,
		},
	}
	return validationRow
}

func metaDataValidation(currResp *types.DescribeDomainResponse, newResp *types.DescribeDomainResponse) (bool, string) {
	if !reflect.DeepEqual(currResp.Configuration, newResp.Configuration) {
		return false, "mismatched DomainConfiguration"
	}

	if currResp.DomainInfo.OwnerEmail != newResp.DomainInfo.OwnerEmail {
		return false, "mismatched OwnerEmail"
	}
	return true, ""
}

func (d *domainMigrationCLIImpl) DomainWorkFlowCheck(c *cli.Context) DomainMigrationRow {
	countWorkFlows := d.countLongRunningWorkflow(c)
	check := countWorkFlows == 0
	return DomainMigrationRow{
		ValidationCheck:  "Workflow Check",
		ValidationResult: check,
		ValidationDetails: ValidationDetails{
			LongRunningWorkFlowNum: &countWorkFlows,
		},
	}
}

func (d *domainMigrationCLIImpl) countLongRunningWorkflow(c *cli.Context) int {
	domain := c.GlobalString(FlagDomain)
	thresholdOfLongRunning := time.Now().Add(-longRunningDuration)
	request := &types.CountWorkflowExecutionsRequest{
		Domain: domain,
		Query:  "CloseTime=missing AND StartTime < " + strconv.FormatInt(thresholdOfLongRunning.UnixNano(), 10),
	}
	ctx, cancel := newContextForLongPoll(c)
	defer cancel()
	response, err := d.frontendClient.CountWorkflowExecutions(ctx, request)
	if err != nil {
		ErrorAndExit("Failed to count workflow.", err)
	}
	return int(response.GetCount())
}

func (d *domainMigrationCLIImpl) SearchAttributesChecker(c *cli.Context) DomainMigrationRow {
	ctx, cancel := newContext(c)
	defer cancel()

	// getting user provided search attributes
	searchAttributes := c.StringSlice(FlagSearchAttribute)
	if len(searchAttributes) == 0 {
		return DomainMigrationRow{
			ValidationCheck:  "Search Attributes Check",
			ValidationResult: true,
		}
	}

	// Parse the provided search attributes into a map[string]IndexValueType
	requiredAttributes := make(map[string]types.IndexedValueType)
	for _, attr := range searchAttributes {
		parts := strings.SplitN(attr, ":", 2)
		if len(parts) != 2 {
			ErrorAndExit(fmt.Sprintf("Invalid search attribute format: %s", attr), nil)
		}
		key, valueType := parts[0], parts[1]
		ivt, err := parseIndexedValueType(valueType)
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Invalid search attribute type for %s: %s", key, valueType), err)
		}
		requiredAttributes[key] = ivt
	}

	// getting search attributes for current domain
	currentSearchAttributes, err := d.frontendClient.GetSearchAttributes(ctx)
	if err != nil {
		ErrorAndExit("Unable to get search attributes for current domain.", err)
	}

	// getting search attributes for new domain
	destinationSearchAttributes, err := d.destinationClient.GetSearchAttributes(ctx)
	if err != nil {
		ErrorAndExit("Unable to get search attributes for new domain.", err)
	}

	currentSearchAttrs := currentSearchAttributes.Keys
	destinationSearchAttrs := destinationSearchAttributes.Keys

	// checking to see if search attributes exist
	missingInCurrent := findMissingAttributes(requiredAttributes, currentSearchAttrs)
	missingInNew := findMissingAttributes(requiredAttributes, destinationSearchAttrs)

	validationResult := len(missingInCurrent) == 0 && len(missingInNew) == 0

	validationRow := DomainMigrationRow{
		ValidationCheck:  "Search Attributes Check",
		ValidationResult: validationResult,
		ValidationDetails: ValidationDetails{
			MissingCurrSearchAttributes: missingInCurrent,
			MissingNewSearchAttributes:  missingInNew,
		},
	}

	return validationRow
}

// helper to parse types.IndexedValueType from string
func parseIndexedValueType(valueType string) (types.IndexedValueType, error) {
	var result types.IndexedValueType
	valueTypeBytes := []byte(valueType)
	if err := result.UnmarshalText(valueTypeBytes); err != nil {
		return 0, err
	}
	return result, nil
}

// finds missing attributed in a map of existing attributed based on required attributes
func findMissingAttributes(requiredAttributes map[string]types.IndexedValueType, existingAttributes map[string]types.IndexedValueType) []string {
	missingAttributes := make([]string, 0)
	for key, requiredType := range requiredAttributes {
		existingType, ok := existingAttributes[key]
		if !ok || existingType != requiredType {
			// construct the key:type string format
			attr := fmt.Sprintf("%s:%s", key, requiredType)
			missingAttributes = append(missingAttributes, attr)
		}
	}
	return missingAttributes
}

func (d *domainMigrationCLIImpl) DynamicConfigCheck(c *cli.Context) DomainMigrationRow {
	var mismatchedConfigs []MismatchedDynamicConfig
	check := true

	resp := dynamicconfig.ListAllProductionKeys()

	currDomain := c.GlobalString(FlagDomain)
	newDomain := c.String(FlagDestinationDomain)

	ctx, cancel := newContext(c)
	defer cancel()

	currentDomainID := getDomainID(ctx, currDomain, d.frontendClient)
	destinationDomainID := getDomainID(ctx, newDomain, d.destinationClient)
	if currentDomainID == "" {
		ErrorAndExit("Failed to get domainID for the current domain.", nil)
	}

	if destinationDomainID == "" {
		ErrorAndExit("Failed to get domainID for the destination domain.", nil)
	}

	for _, configKey := range resp {
		if len(configKey.Filters()) == 1 && configKey.Filters()[0] == dc.DomainName {
			// Validate dynamic configs with only domainName filter
			currRequest := dynamicconfig.ToGetDynamicConfigFilterRequest(configKey.String(), []dynamicconfig.FilterOption{
				dynamicconfig.DomainFilter(currDomain),
			})

			newRequest := dynamicconfig.ToGetDynamicConfigFilterRequest(configKey.String(), []dynamicconfig.FilterOption{
				dynamicconfig.DomainFilter(newDomain),
			})

			currResp, err := d.frontendAdminClient.GetDynamicConfig(ctx, currRequest)
			if err != nil {
				// empty to indicate N/A
				currResp = emptyGetDynamicConfigRequest
			}
			newResp, err := d.destinationAdminClient.GetDynamicConfig(ctx, newRequest)
			if err != nil {
				// empty to indicate N/A
				newResp = emptyGetDynamicConfigRequest
			}

			if !reflect.DeepEqual(currResp.Value, newResp.Value) {
				check = false
				mismatchedConfigs = append(mismatchedConfigs, MismatchedDynamicConfig{
					Key: configKey,
					CurrValues: []*types.DynamicConfigValue{
						toDynamicConfigValue(currResp.Value, map[dc.Filter]interface{}{
							dynamicconfig.DomainName: currDomain,
						}),
					},
					NewValues: []*types.DynamicConfigValue{
						toDynamicConfigValue(newResp.Value, map[dc.Filter]interface{}{
							dynamicconfig.DomainName: newDomain,
						}),
					},
				})
			}

		} else if len(configKey.Filters()) == 1 && configKey.Filters()[0] == dc.DomainID {
			// Validate dynamic configs with only domainID filter
			currRequest := dynamicconfig.ToGetDynamicConfigFilterRequest(configKey.String(), []dynamicconfig.FilterOption{
				dynamicconfig.DomainIDFilter(currentDomainID),
			})

			newRequest := dynamicconfig.ToGetDynamicConfigFilterRequest(configKey.String(), []dynamicconfig.FilterOption{
				dynamicconfig.DomainIDFilter(destinationDomainID),
			})

			currResp, err := d.frontendAdminClient.GetDynamicConfig(ctx, currRequest)
			if err != nil {
				// empty to indicate N/A
				currResp = emptyGetDynamicConfigRequest
			}
			newResp, err := d.destinationAdminClient.GetDynamicConfig(ctx, newRequest)
			if err != nil {
				// empty to indicate N/A
				newResp = emptyGetDynamicConfigRequest
			}

			if !reflect.DeepEqual(currResp.Value, newResp.Value) {
				check = false
				mismatchedConfigs = append(mismatchedConfigs, MismatchedDynamicConfig{
					Key: configKey,
					CurrValues: []*types.DynamicConfigValue{
						toDynamicConfigValue(currResp.Value, map[dc.Filter]interface{}{
							dynamicconfig.DomainID: currentDomainID,
						}),
					},
					NewValues: []*types.DynamicConfigValue{
						toDynamicConfigValue(newResp.Value, map[dc.Filter]interface{}{
							dynamicconfig.DomainID: destinationDomainID,
						}),
					},
				})
			}

		} else if containsFilter(configKey, dc.DomainName.String()) && containsFilter(configKey, dc.TaskListName.String()) {
			// Validate dynamic configs with only domainName and TaskList filters
			taskLists := c.StringSlice(FlagTaskList)
			var mismatchedCurValues []*types.DynamicConfigValue
			var mismatchedNewValues []*types.DynamicConfigValue
			for _, taskList := range taskLists {

				currRequest := dynamicconfig.ToGetDynamicConfigFilterRequest(configKey.String(), []dynamicconfig.FilterOption{
					dynamicconfig.DomainFilter(currDomain),
					dynamicconfig.TaskListFilter(taskList),
				})

				newRequest := dynamicconfig.ToGetDynamicConfigFilterRequest(configKey.String(), []dynamicconfig.FilterOption{
					dynamicconfig.DomainFilter(newDomain),
					dynamicconfig.TaskListFilter(taskList),
				})

				currResp, err := d.frontendAdminClient.GetDynamicConfig(ctx, currRequest)
				if err != nil {
					// empty to indicate N/A
					currResp = emptyGetDynamicConfigRequest
				}
				newResp, err := d.destinationAdminClient.GetDynamicConfig(ctx, newRequest)
				if err != nil {
					// empty to indicate N/A
					newResp = emptyGetDynamicConfigRequest
				}

				if !reflect.DeepEqual(currResp.Value, newResp.Value) {
					check = false
					mismatchedCurValues = append(mismatchedCurValues, toDynamicConfigValue(currResp.Value, map[dc.Filter]interface{}{
						dynamicconfig.DomainName:   currDomain,
						dynamicconfig.TaskListName: taskLists,
					}))
					mismatchedNewValues = append(mismatchedNewValues, toDynamicConfigValue(newResp.Value, map[dc.Filter]interface{}{
						dynamicconfig.DomainName:   newDomain,
						dynamicconfig.TaskListName: taskLists,
					}))

				}
			}
			if len(mismatchedCurValues) > 0 && len(mismatchedNewValues) > 0 {
				mismatchedConfigs = append(mismatchedConfigs, MismatchedDynamicConfig{
					Key:        configKey,
					CurrValues: mismatchedCurValues,
					NewValues:  mismatchedNewValues,
				})
			}
		}
	}

	validationRow := DomainMigrationRow{
		ValidationCheck:  "Dynamic Config Check",
		ValidationResult: check,
		ValidationDetails: ValidationDetails{
			MismatchedDynamicConfig: mismatchedConfigs,
		},
	}

	return validationRow
}

func getDomainID(c context.Context, domain string, client frontend.Client) string {
	resp, err := client.DescribeDomain(c, &types.DescribeDomainRequest{Name: &domain})
	if err != nil {
		ErrorAndExit("Failed to describe domain.", err)
	}

	return resp.DomainInfo.GetUUID()
}

func valueToDataBlob(value interface{}) *types.DataBlob {
	if value == nil {
		return nil
	}
	// No need to handle error as this is a private helper method
	// where the correct value will always be passed regardless
	data, _ := json.Marshal(value)

	return &types.DataBlob{
		EncodingType: types.EncodingTypeJSON.Ptr(),
		Data:         data,
	}
}

func toDynamicConfigValue(value *types.DataBlob, filterMaps map[dynamicconfig.Filter]interface{}) *types.DynamicConfigValue {
	var configFilters []*types.DynamicConfigFilter
	for filter, filterValue := range filterMaps {
		configFilters = append(configFilters, &types.DynamicConfigFilter{
			Name:  filter.String(),
			Value: valueToDataBlob(filterValue),
		})
	}

	return &types.DynamicConfigValue{
		Value:   value,
		Filters: configFilters,
	}
}

func containsFilter(key dynamicconfig.Key, value string) bool {
	filters := key.Filters()
	for _, filter := range filters {
		if filter.String() == value {
			return true
		}
	}
	return false
}
