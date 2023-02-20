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

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common/types"
)

type cliEntry struct {
	Name         string
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Values       []*cliValue
}

type cliValue struct {
	Value   interface{}
	Filters []*cliFilter
}

type cliFilter struct {
	Name  string
	Value interface{}
}

// AdminGetDynamicConfig gets value of specified dynamic config parameter matching specified filter
func AdminGetDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := getRequiredOption(c, FlagDynamicConfigName)
	filters := c.StringSlice(FlagDynamicConfigFilter)

	ctx, cancel := newContext(c)
	defer cancel()

	if len(filters) == 0 {
		req := &types.ListDynamicConfigRequest{
			ConfigName: dcName,
		}

		val, err := adminClient.ListDynamicConfig(ctx, req)
		if err != nil {
			ErrorAndExit("Failed to get dynamic config value(s)", err)
		}

		if val == nil || val.Entries == nil || len(val.Entries) == 0 {
			fmt.Printf("No dynamic config values stored to list.\n")
		} else {
			cliEntries := make([]*cliEntry, 0, len(val.Entries))
			for _, dcEntry := range val.Entries {
				cliEntry, err := convertToInputEntry(dcEntry)
				if err != nil {
					fmt.Printf("Cannot parse list response.\n")
				}
				cliEntries = append(cliEntries, cliEntry)
			}
			prettyPrintJSONObject(cliEntries)
		}
	} else {
		parsedFilters, err := parseInputFilterArray(filters)
		if err != nil {
			ErrorAndExit("Failed to parse input filter array", err)
		}

		req := &types.GetDynamicConfigRequest{
			ConfigName: dcName,
			Filters:    parsedFilters,
		}

		val, err := adminClient.GetDynamicConfig(ctx, req)
		if err != nil {
			ErrorAndExit("Failed to get dynamic config value", err)
		}

		var umVal interface{}
		err = json.Unmarshal(val.Value.Data, &umVal)
		if err != nil {
			ErrorAndExit("Failed to unmarshal response", err)
		}

		if umVal == nil {
			fmt.Printf("No values stored for specified dynamic config.\n")
		} else {
			prettyPrintJSONObject(umVal)
		}
	}
}

// AdminUpdateDynamicConfig updates specified dynamic config parameter with specified values
func AdminUpdateDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := getRequiredOption(c, FlagDynamicConfigName)
	dcValues := c.StringSlice(FlagDynamicConfigValue)

	ctx, cancel := newContext(c)
	defer cancel()

	var parsedValues []*types.DynamicConfigValue

	if dcValues != nil {
		parsedValues = make([]*types.DynamicConfigValue, 0, len(dcValues))

		for _, valueString := range dcValues {
			var parsedInputValue *cliValue
			err := json.Unmarshal([]byte(valueString), &parsedInputValue)
			if err != nil {
				ErrorAndExit("Unable to unmarshal value to inputValue", err)
			}
			parsedValue, err := convertFromInputValue(parsedInputValue)
			if err != nil {
				ErrorAndExit("Unable to convert from inputValue to DynamicConfigValue", err)
			}
			parsedValues = append(parsedValues, parsedValue)
		}
	} else {
		parsedValues = nil
	}

	req := &types.UpdateDynamicConfigRequest{
		ConfigName:   dcName,
		ConfigValues: parsedValues,
	}

	err := adminClient.UpdateDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to update dynamic config value", err)
	}
	fmt.Printf("Dynamic Config %q updated with %s \n", dcName, dcValues)
}

// AdminRestoreDynamicConfig removes values of specified dynamic config parameter matching specified filter
func AdminRestoreDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := getRequiredOption(c, FlagDynamicConfigName)
	filters := c.StringSlice(FlagDynamicConfigFilter)

	ctx, cancel := newContext(c)
	defer cancel()

	parsedFilters, err := parseInputFilterArray(filters)
	if err != nil {
		ErrorAndExit("Failed to parse input filter array", err)
	}

	req := &types.RestoreDynamicConfigRequest{
		ConfigName: dcName,
		Filters:    parsedFilters,
	}

	err = adminClient.RestoreDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to restore dynamic config value", err)
	}
	fmt.Printf("Dynamic Config %q restored\n", dcName)
}

// AdminListDynamicConfig lists all values associated with specified dynamic config parameter or all values for all dc parameter if none is specified.
func AdminListDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.ListDynamicConfigRequest{
		ConfigName: "", // empty string means all config values
	}

	val, err := adminClient.ListDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to list dynamic config value(s)", err)
	}

	if val == nil || val.Entries == nil || len(val.Entries) == 0 {
		fmt.Printf("No dynamic config values stored to list.\n")
	} else {
		cliEntries := make([]*cliEntry, 0, len(val.Entries))
		for _, dcEntry := range val.Entries {
			cliEntry, err := convertToInputEntry(dcEntry)
			if err != nil {
				fmt.Printf("Cannot parse list response.\n")
			}
			cliEntries = append(cliEntries, cliEntry)
		}
		prettyPrintJSONObject(cliEntries)
	}
}

func convertToInputEntry(dcEntry *types.DynamicConfigEntry) (*cliEntry, error) {
	newValues := make([]*cliValue, 0, len(dcEntry.Values))
	for _, value := range dcEntry.Values {
		newValue, err := convertToInputValue(value)
		if err != nil {
			return nil, err
		}
		newValues = append(newValues, newValue)
	}
	return &cliEntry{
		Name:   dcEntry.Name,
		Values: newValues,
	}, nil
}

func convertToInputValue(dcValue *types.DynamicConfigValue) (*cliValue, error) {
	newFilters := make([]*cliFilter, 0, len(dcValue.Filters))
	for _, filter := range dcValue.Filters {
		newFilter, err := convertToInputFilter(filter)
		if err != nil {
			return nil, err
		}
		newFilters = append(newFilters, newFilter)
	}

	var val interface{}
	err := json.Unmarshal(dcValue.Value.Data, &val)
	if err != nil {
		return nil, err
	}

	return &cliValue{
		Value:   val,
		Filters: newFilters,
	}, nil
}

func convertToInputFilter(dcFilter *types.DynamicConfigFilter) (*cliFilter, error) {
	var val interface{}
	err := json.Unmarshal(dcFilter.Value.Data, &val)
	if err != nil {
		return nil, err
	}

	return &cliFilter{
		Name:  dcFilter.Name,
		Value: val,
	}, nil
}

func convertFromInputValue(inputValue *cliValue) (*types.DynamicConfigValue, error) {
	encodedValue, err := json.Marshal(inputValue.Value)
	if err != nil {
		return nil, err
	}

	blob := &types.DataBlob{
		EncodingType: types.EncodingTypeJSON.Ptr(),
		Data:         encodedValue,
	}

	dcFilters := make([]*types.DynamicConfigFilter, 0, len(inputValue.Filters))
	for _, inputFilter := range inputValue.Filters {
		dcFilter, err := convertFromInputFilter(inputFilter)
		if err != nil {
			return nil, err
		}
		dcFilters = append(dcFilters, dcFilter)
	}

	return &types.DynamicConfigValue{
		Value:   blob,
		Filters: dcFilters,
	}, nil
}

func convertFromInputFilter(inputFilter *cliFilter) (*types.DynamicConfigFilter, error) {
	encodedValue, err := json.Marshal(inputFilter.Value)
	if err != nil {
		return nil, err
	}

	return &types.DynamicConfigFilter{
		Name: inputFilter.Name,
		Value: &types.DataBlob{
			EncodingType: types.EncodingTypeJSON.Ptr(),
			Data:         encodedValue,
		},
	}, nil
}

func parseInputFilterArray(inputFilters []string) ([]*types.DynamicConfigFilter, error) {
	var parsedFilters []*types.DynamicConfigFilter

	if len(inputFilters) == 1 && (inputFilters[0] == "" || inputFilters[0] == "{}") {
		parsedFilters = nil
	} else {
		parsedFilters = make([]*types.DynamicConfigFilter, 0, len(inputFilters))

		for _, filterString := range inputFilters {
			var parsedInputFilter *cliFilter
			err := json.Unmarshal([]byte(filterString), &parsedInputFilter)
			if err != nil {
				return nil, err
			}

			filter, err := convertFromInputFilter(parsedInputFilter)
			if err != nil {
				return nil, err
			}

			parsedFilters = append(parsedFilters, filter)
		}
	}

	return parsedFilters, nil
}
