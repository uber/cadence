package dynamicconfig

import (
	"context"
	"fmt"
)

type FilterResult struct {
	// Filters holds any applicable filter-type and the value used.
	// "empty" values will not result in an entry.
	Filters map[Filter]interface{}
	// Current value
	Current interface{}
	// Default value
	Default interface{}
}

type FilterValue struct {
	Filter Filter
	Value  interface{}
}

// GetAllDynamicConfigs will step through every known dynamic config property in this collection,
// for each combination of the passed arguments possible (where applicable, most configs only support a couple),
// and every value will be returned.
//
// This is not every config / every value possible, only ones that have been accessed so far in this process.
// For a more complete list, consider checking logs.
//
// Note that this will potentially result in an enormous number of requests to the dynamic config client.
// As such, you are cautioned to only do this when it won't be dangerous, and to pass a context with a reasonable timeout.
//
// See GetNonDefaultDynamicConfigs for only values that do not match in-code config.
func GetAllDynamicConfigs(
	ctx context.Context,
	c *Collection,
	// one per filter option type.
	// clusterName is ignored because it is inferred from the current runtime environment.
	domainNames []string,
	domainIds []string,
	taskListNames []string,
	taskListTypes []int,
	shardIDs []int,
	workflowIDs []string,
	workflowTypes []string,
) map[Key][]FilterResult {
	return getDynamicConfigs(ctx, c, true, domainNames, domainIds, taskListNames, taskListTypes, shardIDs, workflowIDs, workflowTypes)
}

// GetNonDefaultDynamicConfigs will step through every known dynamic config property,
// for each combination of the passed arguments possible (where applicable, most configs only support a couple),
// and every value that does not match the hardcoded default value will be returned.
//
// This is not every config / every value possible, only ones that have been accessed so far in this process.
// For a more complete list, consider checking logs.
//
// Note that this will potentially result in an enormous number of requests to the dynamic config client.
// As such, you are cautioned to only do this when it won't be dangerous, and to pass a context with a reasonable timeout.
//
// See GetAllDynamicConfigs to retrieve all values.
func GetNonDefaultDynamicConfigs(
	ctx context.Context,
	c *Collection,
	// one per filter option type.
	// clusterName is ignored because it is inferred from the current runtime environment.
	domainNames []string,
	domainIds []string,
	taskListNames []string,
	taskListTypes []int,
	shardIDs []int,
	workflowIDs []string,
	workflowTypes []string,
) map[Key][]FilterResult {
	return getDynamicConfigs(ctx, c, false, domainNames, domainIds, taskListNames, taskListTypes, shardIDs, workflowIDs, workflowTypes)
}

func getDynamicConfigs(
	ctx context.Context,
	c *Collection,
	includeDefaults bool,
	// one per filter option type.
	// clusterName is ignored because it is inferred from the current runtime environment.

	domainNames []string,
	domainIds []string,
	taskListNames []string,
	taskListTypes []int,
	shardIDs []int,
	workflowIDs []string,
	workflowTypes []string,
) map[Key][]FilterResult {

	addUnfilterable := func(r map[Key][]FilterResult, k Key, dv, cv interface{}) {
		if dv != cv || includeDefaults {
			r[k] = []FilterResult{{
				Filters: nil,
				Current: cv,
				Default: dv,
			}}
		}
	}

	appendFiltered := func(list []FilterResult, dv, cv interface{}, filters ...FilterValue) []FilterResult {
		if dv != cv || includeDefaults {
			var fs map[Filter]interface{}
			if len(filters) > 0 {
				m := make(map[Filter]interface{}, len(filters))
				for _, f := range filters {
					if _, ok := m[f.Filter]; !ok {
						panic(fmt.Sprintf("coding error, duplicate filter type: %v", f.Filter))
					}
					m[f.Filter] = f.Value
				}
			}
			list = append(list, FilterResult{
				Filters: fs,
				Current: cv,
				Default: dv,
			})
		}
		return list
	}

	all := c.AllProperties()
	result := make(map[Key][]FilterResult, len(all))
outer:
	for k, v := range all {
		if ctx.Err() != nil {
			break
		}

		var list []FilterResult

		// I'm unsure if generics can reduce duplication here... but it probably isn't worth a lot of effort to try.
		// This only needs to change when we change *kinds* of dynamic config, which is pretty rare.
		switch prop := v.prop.(type) {
		// no recognized argument types == can't filter it in any meaningful / computationally-feasible way.
		// perhaps we can eliminate these one day.
		case PropertyFn:
			addUnfilterable(result, k, v, prop())
		case BoolPropertyFn:
			addUnfilterable(result, k, v, prop())
		case StringPropertyFn:
			addUnfilterable(result, k, v, prop())
		case MapPropertyFn:
			addUnfilterable(result, k, v, prop())
		case IntPropertyFn:
			addUnfilterable(result, k, v, prop())
		case FloatPropertyFn:
			addUnfilterable(result, k, v, prop())
		case DurationPropertyFn:
			addUnfilterable(result, k, v, prop())

		// all shard IDs are treated the same
		case IntPropertyFnWithShardIDFilter:
			for _, s := range shardIDs {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(s), FilterValue{ShardID, s})
			}
		case FloatPropertyFnWithShardIDFilter:
			for _, s := range shardIDs {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(s), FilterValue{ShardID, s})
			}
		case DurationPropertyFnWithShardIDFilter:
			for _, s := range shardIDs {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(s), FilterValue{ShardID, s})
			}

		// all domain names are treated the same
		case IntPropertyFnWithDomainFilter:
			for _, d := range domainNames {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(d), FilterValue{DomainName, d})
			}
		case StringPropertyFnWithDomainFilter:
			for _, d := range domainNames {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(d), FilterValue{DomainName, d})
			}
		case BoolPropertyFnWithDomainFilter:
			for _, d := range domainNames {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(d), FilterValue{DomainName, d})
			}
		case DurationPropertyFnWithDomainFilter:
			for _, d := range domainNames {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(d), FilterValue{DomainName, d})
			}

		// all domain IDs are treated the same
		case DurationPropertyFnWithDomainIDFilter:
			for _, d := range domainIds {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(d), FilterValue{DomainID, d})
			}
		case BoolPropertyFnWithDomainIDFilter:
			for _, d := range domainIds {
				if ctx.Err() != nil {
					break outer
				}
				list = appendFiltered(list, v.defaultValue, prop(d), FilterValue{DomainID, d})
			}

		// all workflow + domain name are treated the same
		case IntPropertyFnWithWorkflowTypeFilter:
			for _, d := range domainNames {
				for _, w := range workflowTypes {
					if ctx.Err() != nil {
						break outer
					}
					list = appendFiltered(list, v.defaultValue, prop(d, w),
						FilterValue{DomainName, d},
						FilterValue{WorkflowType, w})
				}
			}
		case DurationPropertyFnWithWorkflowTypeFilter:
			for _, d := range domainNames {
				for _, w := range workflowTypes {
					if ctx.Err() != nil {
						break outer
					}
					list = appendFiltered(list, v.defaultValue, prop(d, w),
						FilterValue{DomainName, d},
						FilterValue{WorkflowType, w})
				}
			}

		// similar but not quite identical to above
		// kinda odd these three aren't duplicates tbh.
		case BoolPropertyFnWithDomainIDAndWorkflowIDFilter:
			for _, d := range domainIds {
				for _, w := range workflowIDs {
					if ctx.Err() != nil {
						break outer
					}
					list = appendFiltered(list, v.defaultValue, prop(d, w),
						FilterValue{DomainID, d},
						FilterValue{WorkflowID, w})
				}
			}

		// all task-list-infos are treated the same
		case IntPropertyFnWithTaskListInfoFilters:
			for _, d := range domainNames {
				for _, l := range taskListNames {
					for _, t := range taskListTypes {
						if ctx.Err() != nil {
							break outer
						}
						list = appendFiltered(list, v.defaultValue, prop(d, l, t),
							FilterValue{DomainName, d},
							FilterValue{TaskListName, l},
							FilterValue{TaskType, t})
					}
				}
			}
		case DurationPropertyFnWithTaskListInfoFilters:
			for _, d := range domainNames {
				for _, l := range taskListNames {
					for _, t := range taskListTypes {
						if ctx.Err() != nil {
							break outer
						}
						list = appendFiltered(list, v.defaultValue, prop(d, l, t),
							FilterValue{DomainName, d},
							FilterValue{TaskListName, l},
							FilterValue{TaskType, t})
					}
				}
			}
		case BoolPropertyFnWithTaskListInfoFilters:
			for _, d := range domainNames {
				for _, l := range taskListNames {
					for _, t := range taskListTypes {
						if ctx.Err() != nil {
							break outer
						}
						list = appendFiltered(list, v.defaultValue, prop(d, l, t),
							FilterValue{DomainName, d},
							FilterValue{TaskListName, l},
							FilterValue{TaskType, t})
					}
				}
			}
		default:
			// should not make it past CI, but just in case
			result[k] = []FilterResult{{
				Filters: nil,
				Current: fmt.Sprintf("ERROR: unrecognized property type: %T", prop),
				Default: nil,
			}}
		}

		result[k] = list
	}

	return result
}
