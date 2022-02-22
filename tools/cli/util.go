// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/uber/cadence/common/pagination"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common/types"

	"github.com/cristalhq/jwt/v3"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"github.com/valyala/fastjson"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/authorization"
)

// JSONHistorySerializer is used to encode history event in JSON
type JSONHistorySerializer struct{}

// Serialize serializes history.
func (j *JSONHistorySerializer) Serialize(h *types.History) ([]byte, error) {
	return json.Marshal(h.Events)
}

// Deserialize deserializes history
func (j *JSONHistorySerializer) Deserialize(data []byte) (*types.History, error) {
	var events []*types.HistoryEvent
	err := json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}
	return &types.History{Events: events}, nil
}

// GetHistory helper method to iterate over all pages and return complete list of history events
func GetHistory(ctx context.Context, workflowClient frontend.Client, domain, workflowID, runID string) (*types.History, error) {
	events := []*types.HistoryEvent{}
	iterator, err := GetWorkflowHistoryIterator(ctx, workflowClient, domain, workflowID, runID, false, types.HistoryEventFilterTypeAllEvent.Ptr())
	for iterator.HasNext() {
		entity, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		events = append(events, entity.(*types.HistoryEvent))
	}
	history := &types.History{}
	history.Events = events
	return history, err
}

// GetWorkflowHistoryIterator returns a HistoryEvent iterator
func GetWorkflowHistoryIterator(
	ctx context.Context,
	workflowClient frontend.Client,
	domain,
	workflowID,
	runID string,
	isLongPoll bool,
	filterType *types.HistoryEventFilterType,
) (pagination.Iterator, error) {
	paginate := func(ctx context.Context, pageToken pagination.PageToken) (pagination.Page, error) {
		tcCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
		defer cancel()

		var nextPageToken []byte
		if pageToken != nil {
			nextPageToken, _ = pageToken.([]byte)
		}
		request := &types.GetWorkflowExecutionHistoryRequest{
			Domain: domain,
			Execution: &types.WorkflowExecution{
				WorkflowID: workflowID,
				RunID:      runID,
			},
			WaitForNewEvent:        isLongPoll,
			HistoryEventFilterType: filterType,
			NextPageToken:          nextPageToken,
			SkipArchival:           isLongPoll,
		}

		var resp *types.GetWorkflowExecutionHistoryResponse
		var err error
	Loop:
		for {
			resp, err = workflowClient.GetWorkflowExecutionHistory(tcCtx, request)
			if err != nil {
				return pagination.Page{}, err
			}

			if isLongPoll && len(resp.History.Events) == 0 && len(resp.NextPageToken) != 0 {
				request.NextPageToken = resp.NextPageToken
				continue Loop
			}
			break Loop
		}
		entities := make([]pagination.Entity, len(resp.History.Events))
		for i, e := range resp.History.Events {
			entities[i] = e
		}
		var nextToken interface{} = resp.NextPageToken
		if len(resp.NextPageToken) == 0 {
			nextToken = nil
		}
		page := pagination.Page{
			CurrentToken: pageToken,
			NextToken:    nextToken,
			Entities:     entities,
		}
		return page, err
	}
	return pagination.NewIterator(ctx, nil, paginate), nil
}

// HistoryEventToString convert HistoryEvent to string
func HistoryEventToString(e *types.HistoryEvent, printFully bool, maxFieldLength int) string {
	data := getEventAttributes(e)
	return anyToString(data, printFully, maxFieldLength)
}

func anyToString(d interface{}, printFully bool, maxFieldLength int) string {
	v := reflect.ValueOf(d)
	switch v.Kind() {
	case reflect.Ptr:
		return anyToString(v.Elem().Interface(), printFully, maxFieldLength)
	case reflect.Struct:
		var buf bytes.Buffer
		t := reflect.TypeOf(d)
		buf.WriteString("{")
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if f.Kind() == reflect.Invalid {
				continue
			}
			fieldValue := valueToString(f, printFully, maxFieldLength)
			if len(fieldValue) == 0 {
				continue
			}
			if buf.Len() > 1 {
				buf.WriteString(", ")
			}
			fieldName := t.Field(i).Name
			if !isAttributeName(fieldName) {
				if !printFully {
					fieldValue = trimTextAndBreakWords(fieldValue, maxFieldLength)
				} else if maxFieldLength != 0 { // for command run workflow and observe history
					fieldValue = trimText(fieldValue, maxFieldLength)
				}
			}
			if fieldName == "Reason" || fieldName == "Details" || fieldName == "Cause" {
				buf.WriteString(fmt.Sprintf("%s:%s", color.RedString(fieldName), color.MagentaString(fieldValue)))
			} else {
				buf.WriteString(fmt.Sprintf("%s:%s", fieldName, fieldValue))
			}
		}
		buf.WriteString("}")
		return buf.String()
	default:
		return fmt.Sprint(d)
	}
}

func valueToString(v reflect.Value, printFully bool, maxFieldLength int) string {
	switch v.Kind() {
	case reflect.Ptr:
		return valueToString(v.Elem(), printFully, maxFieldLength)
	case reflect.Struct:
		return anyToString(v.Interface(), printFully, maxFieldLength)
	case reflect.Invalid:
		return ""
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			n := string(v.Bytes())
			if n != "" && n[len(n)-1] == '\n' {
				return fmt.Sprintf("[%v]", n[:len(n)-1])
			}
			return fmt.Sprintf("[%v]", n)
		}
		return fmt.Sprintf("[len=%d]", v.Len())
	case reflect.Map:
		str := "map{"
		for i, key := range v.MapKeys() {
			str += key.String() + ":"
			val := v.MapIndex(key)
			switch val.Interface().(type) {
			case []byte:
				str += string(val.Interface().([]byte))
			default:
				str += val.String()
			}
			if i != len(v.MapKeys())-1 {
				str += ", "
			}
		}
		str += "}"
		return str
	default:
		return fmt.Sprint(v.Interface())
	}
}

// limit the maximum length for each field
func trimText(input string, maxFieldLength int) string {
	if len(input) > maxFieldLength {
		input = fmt.Sprintf("%s ... %s", input[:maxFieldLength/2], input[(len(input)-maxFieldLength/2):])
	}
	return input
}

// limit the maximum length for each field, and break long words for table item correctly wrap words
func trimTextAndBreakWords(input string, maxFieldLength int) string {
	input = trimText(input, maxFieldLength)
	return breakLongWords(input, maxWordLength)
}

// long words will make output in table cell looks bad,
// break long text "ltltltltllt..." to "ltlt ltlt lt..." will make use of table autowrap so that output is pretty.
func breakLongWords(input string, maxWordLength int) string {
	if len(input) <= maxWordLength {
		return input
	}

	cnt := 0
	for i := 0; i < len(input); i++ {
		if cnt == maxWordLength {
			cnt = 0
			input = input[:i] + " " + input[i:]
			continue
		}
		cnt++
		if input[i] == ' ' {
			cnt = 0
		}
	}
	return input
}

// ColorEvent takes an event and return string with color
// Event with color mapping rules:
//   Failed - red
//   Timeout - yellow
//   Canceled - magenta
//   Completed - green
//   Started - blue
//   Others - default (white/black)
func ColorEvent(e *types.HistoryEvent) string {
	var data string
	switch e.GetEventType() {
	case types.EventTypeWorkflowExecutionStarted:
		data = color.BlueString(e.EventType.String())

	case types.EventTypeWorkflowExecutionCompleted:
		data = color.GreenString(e.EventType.String())

	case types.EventTypeWorkflowExecutionFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeWorkflowExecutionTimedOut:
		data = color.YellowString(e.EventType.String())

	case types.EventTypeDecisionTaskScheduled:
		data = e.EventType.String()

	case types.EventTypeDecisionTaskStarted:
		data = e.EventType.String()

	case types.EventTypeDecisionTaskCompleted:
		data = e.EventType.String()

	case types.EventTypeDecisionTaskTimedOut:
		data = color.YellowString(e.EventType.String())

	case types.EventTypeActivityTaskScheduled:
		data = e.EventType.String()

	case types.EventTypeActivityTaskStarted:
		data = e.EventType.String()

	case types.EventTypeActivityTaskCompleted:
		data = e.EventType.String()

	case types.EventTypeActivityTaskFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeActivityTaskTimedOut:
		data = color.YellowString(e.EventType.String())

	case types.EventTypeActivityTaskCancelRequested:
		data = e.EventType.String()

	case types.EventTypeRequestCancelActivityTaskFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeActivityTaskCanceled:
		data = e.EventType.String()

	case types.EventTypeTimerStarted:
		data = e.EventType.String()

	case types.EventTypeTimerFired:
		data = e.EventType.String()

	case types.EventTypeCancelTimerFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeTimerCanceled:
		data = color.MagentaString(e.EventType.String())

	case types.EventTypeWorkflowExecutionCancelRequested:
		data = e.EventType.String()

	case types.EventTypeWorkflowExecutionCanceled:
		data = color.MagentaString(e.EventType.String())

	case types.EventTypeRequestCancelExternalWorkflowExecutionInitiated:
		data = e.EventType.String()

	case types.EventTypeRequestCancelExternalWorkflowExecutionFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeExternalWorkflowExecutionCancelRequested:
		data = e.EventType.String()

	case types.EventTypeMarkerRecorded:
		data = e.EventType.String()

	case types.EventTypeWorkflowExecutionSignaled:
		data = e.EventType.String()

	case types.EventTypeWorkflowExecutionTerminated:
		data = e.EventType.String()

	case types.EventTypeWorkflowExecutionContinuedAsNew:
		data = e.EventType.String()

	case types.EventTypeStartChildWorkflowExecutionInitiated:
		data = e.EventType.String()

	case types.EventTypeStartChildWorkflowExecutionFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeChildWorkflowExecutionStarted:
		data = color.BlueString(e.EventType.String())

	case types.EventTypeChildWorkflowExecutionCompleted:
		data = color.GreenString(e.EventType.String())

	case types.EventTypeChildWorkflowExecutionFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeChildWorkflowExecutionCanceled:
		data = color.MagentaString(e.EventType.String())

	case types.EventTypeChildWorkflowExecutionTimedOut:
		data = color.YellowString(e.EventType.String())

	case types.EventTypeChildWorkflowExecutionTerminated:
		data = e.EventType.String()

	case types.EventTypeSignalExternalWorkflowExecutionInitiated:
		data = e.EventType.String()

	case types.EventTypeSignalExternalWorkflowExecutionFailed:
		data = color.RedString(e.EventType.String())

	case types.EventTypeExternalWorkflowExecutionSignaled:
		data = e.EventType.String()

	case types.EventTypeUpsertWorkflowSearchAttributes:
		data = e.EventType.String()

	default:
		data = e.EventType.String()
	}
	return data
}

func getEventAttributes(e *types.HistoryEvent) interface{} {
	var data interface{}
	switch e.GetEventType() {
	case types.EventTypeWorkflowExecutionStarted:
		data = e.WorkflowExecutionStartedEventAttributes

	case types.EventTypeWorkflowExecutionCompleted:
		data = e.WorkflowExecutionCompletedEventAttributes

	case types.EventTypeWorkflowExecutionFailed:
		data = e.WorkflowExecutionFailedEventAttributes

	case types.EventTypeWorkflowExecutionTimedOut:
		data = e.WorkflowExecutionTimedOutEventAttributes

	case types.EventTypeDecisionTaskScheduled:
		data = e.DecisionTaskScheduledEventAttributes

	case types.EventTypeDecisionTaskStarted:
		data = e.DecisionTaskStartedEventAttributes

	case types.EventTypeDecisionTaskCompleted:
		data = e.DecisionTaskCompletedEventAttributes

	case types.EventTypeDecisionTaskTimedOut:
		data = e.DecisionTaskTimedOutEventAttributes

	case types.EventTypeActivityTaskScheduled:
		data = e.ActivityTaskScheduledEventAttributes

	case types.EventTypeActivityTaskStarted:
		data = e.ActivityTaskStartedEventAttributes

	case types.EventTypeActivityTaskCompleted:
		data = e.ActivityTaskCompletedEventAttributes

	case types.EventTypeActivityTaskFailed:
		data = e.ActivityTaskFailedEventAttributes

	case types.EventTypeActivityTaskTimedOut:
		data = e.ActivityTaskTimedOutEventAttributes

	case types.EventTypeActivityTaskCancelRequested:
		data = e.ActivityTaskCancelRequestedEventAttributes

	case types.EventTypeRequestCancelActivityTaskFailed:
		data = e.RequestCancelActivityTaskFailedEventAttributes

	case types.EventTypeActivityTaskCanceled:
		data = e.ActivityTaskCanceledEventAttributes

	case types.EventTypeTimerStarted:
		data = e.TimerStartedEventAttributes

	case types.EventTypeTimerFired:
		data = e.TimerFiredEventAttributes

	case types.EventTypeCancelTimerFailed:
		data = e.CancelTimerFailedEventAttributes

	case types.EventTypeTimerCanceled:
		data = e.TimerCanceledEventAttributes

	case types.EventTypeWorkflowExecutionCancelRequested:
		data = e.WorkflowExecutionCancelRequestedEventAttributes

	case types.EventTypeWorkflowExecutionCanceled:
		data = e.WorkflowExecutionCanceledEventAttributes

	case types.EventTypeRequestCancelExternalWorkflowExecutionInitiated:
		data = e.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes

	case types.EventTypeRequestCancelExternalWorkflowExecutionFailed:
		data = e.RequestCancelExternalWorkflowExecutionFailedEventAttributes

	case types.EventTypeExternalWorkflowExecutionCancelRequested:
		data = e.ExternalWorkflowExecutionCancelRequestedEventAttributes

	case types.EventTypeMarkerRecorded:
		data = e.MarkerRecordedEventAttributes

	case types.EventTypeWorkflowExecutionSignaled:
		data = e.WorkflowExecutionSignaledEventAttributes

	case types.EventTypeWorkflowExecutionTerminated:
		data = e.WorkflowExecutionTerminatedEventAttributes

	case types.EventTypeWorkflowExecutionContinuedAsNew:
		data = e.WorkflowExecutionContinuedAsNewEventAttributes

	case types.EventTypeStartChildWorkflowExecutionInitiated:
		data = e.StartChildWorkflowExecutionInitiatedEventAttributes

	case types.EventTypeStartChildWorkflowExecutionFailed:
		data = e.StartChildWorkflowExecutionFailedEventAttributes

	case types.EventTypeChildWorkflowExecutionStarted:
		data = e.ChildWorkflowExecutionStartedEventAttributes

	case types.EventTypeChildWorkflowExecutionCompleted:
		data = e.ChildWorkflowExecutionCompletedEventAttributes

	case types.EventTypeChildWorkflowExecutionFailed:
		data = e.ChildWorkflowExecutionFailedEventAttributes

	case types.EventTypeChildWorkflowExecutionCanceled:
		data = e.ChildWorkflowExecutionCanceledEventAttributes

	case types.EventTypeChildWorkflowExecutionTimedOut:
		data = e.ChildWorkflowExecutionTimedOutEventAttributes

	case types.EventTypeChildWorkflowExecutionTerminated:
		data = e.ChildWorkflowExecutionTerminatedEventAttributes

	case types.EventTypeSignalExternalWorkflowExecutionInitiated:
		data = e.SignalExternalWorkflowExecutionInitiatedEventAttributes

	case types.EventTypeSignalExternalWorkflowExecutionFailed:
		data = e.SignalExternalWorkflowExecutionFailedEventAttributes

	case types.EventTypeExternalWorkflowExecutionSignaled:
		data = e.ExternalWorkflowExecutionSignaledEventAttributes

	default:
		data = e
	}
	return data
}

func isAttributeName(name string) bool {
	for i := types.EventType(0); i < types.EventType(40); i++ {
		if name == i.String()+"EventAttributes" {
			return true
		}
	}
	return false
}

func getCurrentUserFromEnv() string {
	for _, n := range envKeysForUserName {
		if len(os.Getenv(n)) > 0 {
			return os.Getenv(n)
		}
	}
	return "unkown"
}

func prettyPrintJSONObject(o interface{}) {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		fmt.Printf("Error when try to print pretty: %v\n", err)
		fmt.Println(o)
	}
	os.Stdout.Write(b)
	fmt.Println()
}

func mapKeysToArray(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func mapToString(m map[string]string, sep string) string {
	kv := make([]string, 0, len(m))
	for key, value := range m {
		kv = append(kv, key+": "+value)
	}

	return strings.Join(kv, sep)
}

func intSliceToSet(s []int) map[int]struct{} {
	var ret = make(map[int]struct{}, len(s))
	for _, v := range s {
		ret[v] = struct{}{}
	}
	return ret
}

func printError(msg string, err error) {
	if err != nil {
		fmt.Printf("%s %s\n%s %+v\n", colorRed("Error:"), msg, colorMagenta("Error Details:"), err)
		if os.Getenv(showErrorStackEnv) != `` {
			fmt.Printf("Stack trace:\n")
			debug.PrintStack()
		} else {
			fmt.Printf("('export %s=1' to see stack traces)\n", showErrorStackEnv)
		}
	} else {
		fmt.Printf("%s %s\n", colorRed("Error:"), msg)
	}
}

// ErrorAndExit print easy to understand error msg first then error detail in a new line
func ErrorAndExit(msg string, err error) {
	printError(msg, err)
	osExit(1)
}

func getWorkflowClient(c *cli.Context) frontend.Client {
	return cFactory.ServerFrontendClient(c)
}

func getRequiredOption(c *cli.Context, optionName string) string {
	value := c.String(optionName)
	if len(value) == 0 {
		ErrorAndExit(fmt.Sprintf("Option %s is required", optionName), nil)
	}
	return value
}

func getRequiredInt64Option(c *cli.Context, optionName string) int64 {
	if !c.IsSet(optionName) {
		ErrorAndExit(fmt.Sprintf("Option %s is required", optionName), nil)
	}
	return c.Int64(optionName)
}

func getRequiredIntOption(c *cli.Context, optionName string) int {
	if !c.IsSet(optionName) {
		ErrorAndExit(fmt.Sprintf("Option %s is required", optionName), nil)
	}
	return c.Int(optionName)
}

func getRequiredGlobalOption(c *cli.Context, optionName string) string {
	value := c.GlobalString(optionName)
	if len(value) == 0 {
		ErrorAndExit(fmt.Sprintf("Global option %s is required", optionName), nil)
	}
	return value
}

func timestampPtrToStringPtr(unixNanoPtr *int64, onlyTime bool) *string {
	if unixNanoPtr == nil {
		return nil
	}
	return common.StringPtr(convertTime(*unixNanoPtr, onlyTime))
}

func convertTime(unixNano int64, onlyTime bool) string {
	t := time.Unix(0, unixNano)
	var result string
	if onlyTime {
		result = t.Format(defaultTimeFormat)
	} else {
		result = t.Format(defaultDateTimeFormat)
	}
	return result
}

func parseTime(timeStr string, defaultValue int64) int64 {
	if len(timeStr) == 0 {
		return defaultValue
	}

	// try to parse
	parsedTime, err := time.Parse(defaultDateTimeFormat, timeStr)
	if err == nil {
		return parsedTime.UnixNano()
	}

	// treat as raw time
	resultValue, err := strconv.ParseInt(timeStr, 10, 64)
	if err == nil {
		return resultValue
	}

	// treat as time range format
	parsedTime, err = parseTimeRange(timeStr)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Cannot parse time '%s', use UTC format '2006-01-02T15:04:05Z', "+
			"time range or raw UnixNano directly. See help for more details.", timeStr), err)
	}
	return parsedTime.UnixNano()
}

// parseTimeRange parses a given time duration string (in format X<time-duration>) and
// returns parsed timestamp given that duration in the past from current time.
// All valid values must contain a number followed by a time-duration, from the following list (long form/short form):
// - second/s
// - minute/m
// - hour/h
// - day/d
// - week/w
// - month/M
// - year/y
// For example, possible input values, and their result:
// - "3d" or "3day" --> three days --> time.Now().Add(-3 * 24 * time.Hour)
// - "2m" or "2minute" --> two minutes --> time.Now().Add(-2 * time.Minute)
// - "1w" or "1week" --> one week --> time.Now().Add(-7 * 24 * time.Hour)
// - "30s" or "30second" --> thirty seconds --> time.Now().Add(-30 * time.Second)
// Note: Duration strings are case-sensitive, and should be used as mentioned above only.
// Limitation: Value of numerical multiplier, X should be in b/w 0 - 1e6 (1 million), boundary values excluded i.e.
// 0 < X < 1e6. Also, the maximum time in the past can be 1 January 1970 00:00:00 UTC (epoch time),
// so giving "1000y" will result in epoch time.
func parseTimeRange(timeRange string) (time.Time, error) {
	match, err := regexp.MatchString(defaultDateTimeRangeShortRE, timeRange)
	if !match { // fallback on to check if it's of longer notation
		match, err = regexp.MatchString(defaultDateTimeRangeLongRE, timeRange)
	}
	if err != nil {
		return time.Time{}, err
	}

	re, _ := regexp.Compile(defaultDateTimeRangeNum)
	idx := re.FindStringSubmatchIndex(timeRange)
	if idx == nil {
		return time.Time{}, fmt.Errorf("cannot parse timeRange %s", timeRange)
	}

	num, err := strconv.Atoi(timeRange[idx[0]:idx[1]])
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse timeRange %s", timeRange)
	}
	if num >= 1e6 {
		return time.Time{}, fmt.Errorf("invalid time-duation multiplier %d, allowed range is 0 < multiplier < 1000000", num)
	}

	dur, err := parseTimeDuration(timeRange[idx[1]:])
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse timeRange %s", timeRange)
	}

	res := time.Now().Add(time.Duration(-num) * dur) // using server's local timezone
	epochTime := time.Unix(0, 0)
	if res.Before(epochTime) {
		res = epochTime
	}
	return res, nil
}

func parseSingleTs(ts string) (time.Time, error) {
	var tsOut time.Time
	var err error
	formats := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02", "2006-01-02T15:04:05+0700", time.RFC3339}
	for _, format := range formats {
		if tsOut, err = time.Parse(format, ts); err == nil {
			return tsOut, err
		}
	}
	return tsOut, err
}

// parseTimeDuration parses the given time duration in either short or long convention
// and returns the time.Duration
// Valid values (long notation/short notation):
// - second/s
// - minute/m
// - hour/h
// - day/d
// - week/w
// - month/M
// - year/y
// NOTE: the input "duration" is case-sensitive
func parseTimeDuration(duration string) (dur time.Duration, err error) {
	switch duration {
	case "s", "second":
		dur = time.Second
	case "m", "minute":
		dur = time.Minute
	case "h", "hour":
		dur = time.Hour
	case "d", "day":
		dur = day
	case "w", "week":
		dur = week
	case "M", "month":
		dur = month
	case "y", "year":
		dur = year
	default:
		err = fmt.Errorf("unknown time duration %s", duration)
	}
	return
}

func strToTaskListType(str string) types.TaskListType {
	if strings.ToLower(str) == "activity" {
		return types.TaskListTypeActivity
	}
	return types.TaskListTypeDecision
}

func getCliIdentity() string {
	return fmt.Sprintf("cadence-cli@%s", getHostName())
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "UnKnown"
	}
	return hostName
}

func processJWTFlags(ctx context.Context, cliCtx *cli.Context) context.Context {
	path := getJWTPrivateKey(cliCtx)
	t := getJWT(cliCtx)
	var token string

	if t != "" {
		token = t
	} else if path != "" {
		createdToken, err := createJWT(path)
		if err != nil {
			ErrorAndExit("Error creating JWT token", err)
		}
		token = *createdToken
	}

	ctx = context.WithValue(ctx, CtxKeyJWT, token)
	return ctx
}

func populateContextFromCLIContext(ctx context.Context, cliCtx *cli.Context) context.Context {
	ctx = processJWTFlags(ctx, cliCtx)
	return ctx
}

func newContext(c *cli.Context) (context.Context, context.CancelFunc) {
	contextTimeout := defaultContextTimeout
	if c.GlobalInt(FlagContextTimeout) > 0 {
		contextTimeout = time.Duration(c.GlobalInt(FlagContextTimeout)) * time.Second
	}
	ctx := populateContextFromCLIContext(context.Background(), c)
	return context.WithTimeout(ctx, contextTimeout)
}

func newContextForLongPoll(c *cli.Context) (context.Context, context.CancelFunc) {
	contextTimeout := defaultContextTimeoutForLongPoll
	if c.GlobalIsSet(FlagContextTimeout) {
		contextTimeout = time.Duration(c.GlobalInt(FlagContextTimeout)) * time.Second
	}
	return context.WithTimeout(context.Background(), contextTimeout)
}

func newIndefiniteContext(c *cli.Context) (context.Context, context.CancelFunc) {
	if c.GlobalIsSet(FlagContextTimeout) {
		contextTimeout := time.Duration(c.GlobalInt(FlagContextTimeout)) * time.Second
		return context.WithTimeout(context.Background(), contextTimeout)
	}

	return context.WithCancel(context.Background())
}

// process and validate input provided through cmd or file
func processJSONInput(c *cli.Context) string {
	return processJSONInputHelper(c, jsonTypeInput)
}

// process and validate json
func processJSONInputHelper(c *cli.Context, jType jsonType) string {
	var flagNameOfRawInput string
	var flagNameOfInputFileName string

	switch jType {
	case jsonTypeInput:
		flagNameOfRawInput = FlagInput
		flagNameOfInputFileName = FlagInputFile
	case jsonTypeMemo:
		flagNameOfRawInput = FlagMemo
		flagNameOfInputFileName = FlagMemoFile
	case jsonTypeHeader:
		flagNameOfRawInput = FlagHeaderValue
		flagNameOfInputFileName = FlagHeaderFile
	case jsonTypeSignal:
		flagNameOfRawInput = FlagSignalInput
		flagNameOfInputFileName = FlagSignalInputFile
	default:
		return ""
	}

	var input string
	if c.IsSet(flagNameOfRawInput) {
		input = c.String(flagNameOfRawInput)
	} else if c.IsSet(flagNameOfInputFileName) {
		inputFile := c.String(flagNameOfInputFileName)
		// This method is purely used to parse input from the CLI. The input comes from a trusted user
		// #nosec
		data, err := ioutil.ReadFile(inputFile)
		if err != nil {
			ErrorAndExit("Error reading input file", err)
		}
		input = string(data)
	}
	if input != "" {
		if err := validateJSONs(input); err != nil {
			ErrorAndExit("Input is not valid JSON, or JSONs concatenated with spaces/newlines.", err)
		}
	}
	return input
}

func processMultipleKeys(rawKey, separator string) []string {
	var keys []string
	if strings.TrimSpace(rawKey) != "" {
		keys = strings.Split(rawKey, separator)
	}
	return keys
}

func processMultipleJSONValues(rawValue string) []string {
	var values []string
	var sc fastjson.Scanner
	sc.Init(rawValue)
	for sc.Next() {
		values = append(values, sc.Value().String())
	}
	if err := sc.Error(); err != nil {
		ErrorAndExit("Parse json error.", err)
	}
	return values
}

func mapFromKeysValues(keys, values []string) map[string][]byte {
	fields := map[string][]byte{}
	for i, key := range keys {
		fields[key] = []byte(values[i])
	}
	return fields
}

// validate whether str is a valid json or multi valid json concatenated with spaces/newlines
func validateJSONs(str string) error {
	input := []byte(str)
	dec := json.NewDecoder(bytes.NewReader(input))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return nil // End of input, valid JSON
		}
		if err != nil {
			return err // Invalid input
		}
	}
}

// use parseBool to ensure all BOOL search attributes only be "true" or "false"
func parseBool(str string) (bool, error) {
	switch str {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("not parseable bool value: %s", str)
}

func trimSpace(strs []string) []string {
	result := make([]string, len(strs))
	for i, v := range strs {
		result[i] = strings.TrimSpace(v)
	}
	return result
}

func parseArray(v string) (interface{}, error) {
	if len(v) > 0 && v[0] == '[' && v[len(v)-1] == ']' {
		parsedValues, err := fastjson.Parse(v)
		if err != nil {
			return nil, err
		}
		arr, err := parsedValues.Array()
		if err != nil {
			return nil, err
		}
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			s := item.String()
			if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' { // remove addition quote from json
				s = s[1 : len(s)-1]
				if sTime, err := time.Parse(defaultDateTimeFormat, s); err == nil {
					result[i] = sTime
					continue
				}
			}
			result[i] = s
		}
		return result, nil
	}
	return nil, errors.New("not array")
}

func convertStringToRealType(v string) interface{} {
	var genVal interface{}
	var err error

	if genVal, err = strconv.ParseInt(v, 10, 64); err == nil {

	} else if genVal, err = parseBool(v); err == nil {

	} else if genVal, err = strconv.ParseFloat(v, 64); err == nil {

	} else if genVal, err = time.Parse(defaultDateTimeFormat, v); err == nil {

	} else if genVal, err = parseArray(v); err == nil {

	} else {
		genVal = v
	}

	return genVal
}

func truncate(str string) string {
	if len(str) > maxOutputStringLength {
		return str[:maxOutputStringLength]
	}
	return str
}

// this only works for ANSI terminal, which means remove existing lines won't work if users redirect to file
// ref: https://en.wikipedia.org/wiki/ANSI_escape_code
func removePrevious2LinesFromTerminal() {
	fmt.Printf("\033[1A")
	fmt.Printf("\033[2K")
	fmt.Printf("\033[1A")
	fmt.Printf("\033[2K")
}

func showNextPage() bool {
	fmt.Printf("Press %s to show next page, press %s to quit: ",
		color.GreenString("Enter"), color.RedString("any other key then Enter"))
	var input string
	fmt.Scanln(&input)
	return strings.Trim(input, " ") == ""
}

// prompt will show input msg, then waiting user input y/yes to continue
func prompt(msg string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(msg)
	text, _ := reader.ReadString('\n')
	textLower := strings.ToLower(strings.TrimRight(text, "\n"))
	if textLower != "y" && textLower != "yes" {
		os.Exit(0)
	}
}
func getInputFile(inputFile string) *os.File {
	if len(inputFile) == 0 {
		info, err := os.Stdin.Stat()
		if err != nil {
			ErrorAndExit("Failed to stat stdin file handle", err)
		}
		if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
			fmt.Fprintln(os.Stderr, "Provide a filename or pass data to STDIN")
			os.Exit(1)
		}
		return os.Stdin
	}
	// This code is executed from the CLI. All user input is from a CLI user.
	// #nosec
	f, err := os.Open(inputFile)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Failed to open input file for reading: %v", inputFile), err)
	}
	return f
}

// createJWT defines the logic to create a JWT
func createJWT(keyPath string) (*string, error) {
	claims := authorization.JWTClaims{
		Admin: true,
		Iat:   time.Now().Unix(),
		TTL:   60 * 10,
	}

	privateKey, err := common.LoadRSAPrivateKey(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := jwt.NewSignerRS(jwt.RS256, privateKey)
	if err != nil {
		return nil, err
	}
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claims)
	if token == nil {
		return nil, err
	}
	tokenString := token.String()
	return &tokenString, nil
}

func getWorkflowMemo(input map[string]interface{}) (*types.Memo, error) {
	if input == nil {
		return nil, nil
	}

	memo := make(map[string][]byte)
	for k, v := range input {
		memoBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("encode workflow memo error: %v", err.Error())
		}
		memo[k] = memoBytes
	}
	return &types.Memo{Fields: memo}, nil
}

func serializeSearchAttributes(input map[string]interface{}) (*types.SearchAttributes, error) {
	if input == nil {
		return nil, nil
	}

	attr := make(map[string][]byte)
	for k, v := range input {
		attrBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("encode search attribute [%s] error: %v", k, err)
		}
		attr[k] = attrBytes
	}
	return &types.SearchAttributes{IndexedFields: attr}, nil
}
