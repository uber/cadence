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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/urfave/cli"
	"github.com/valyala/fastjson"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
)

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

// ErrorAndExit print easy to understand error msg first then error detail in a new line
func ErrorAndExit(msg string, err error) {
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
	osExit(1)
}

func getWorkflowClient(c *cli.Context) client.Client {
	domain := getRequiredGlobalOption(c, FlagDomain)
	return client.NewClient(cFactory.ClientFrontendClient(c), domain, &client.Options{})
}

func getRequiredOption(c *cli.Context, optionName string) string {
	value := c.String(optionName)
	if len(value) == 0 {
		ErrorAndExit(fmt.Sprintf("Option %s is required", optionName), nil)
	}
	return value
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
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Cannot parse time '%s', use UTC format '2006-01-02T15:04:05Z' or raw UnixNano directly.", timeStr), err)
	}

	return resultValue
}

func strToTaskListType(str string) s.TaskListType {
	if strings.ToLower(str) == "activity" {
		return s.TaskListTypeActivity
	}
	return s.TaskListTypeDecision
}

func getPtrOrNilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return common.StringPtr(value)
}

func getCliIdentity() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "UnKnown"
	}
	return fmt.Sprintf("cadence-cli@%s", hostName)
}

func newContext(c *cli.Context) (context.Context, context.CancelFunc) {
	contextTimeout := defaultContextTimeout
	if c.GlobalInt(FlagContextTimeout) > 0 {
		contextTimeout = time.Duration(c.GlobalInt(FlagContextTimeout)) * time.Second
	}
	return context.WithTimeout(context.Background(), contextTimeout)
}

func newContextForLongPoll(c *cli.Context) (context.Context, context.CancelFunc) {
	contextTimeout := defaultContextTimeoutForLongPoll
	if c.GlobalIsSet(FlagContextTimeout) {
		contextTimeout = time.Duration(c.GlobalInt(FlagContextTimeout)) * time.Second
	}
	return context.WithTimeout(context.Background(), contextTimeout)
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
	default:
		return ""
	}

	var input string
	if c.IsSet(flagNameOfRawInput) {
		input = c.String(flagNameOfRawInput)
	} else if c.IsSet(flagNameOfInputFileName) {
		inputFile := c.String(flagNameOfInputFileName)
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

func convertStringToRealType(v string) interface{} {
	var genVal interface{}
	var err error

	if genVal, err = strconv.ParseInt(v, 10, 64); err == nil {

	} else if genVal, err = parseBool(v); err == nil {

	} else if genVal, err = strconv.ParseFloat(v, 64); err == nil {

	} else if genVal, err = time.Parse(defaultDateTimeFormat, v); err == nil {

	} else {
		genVal = v
	}

	return genVal
}

func processSearchAttr(c *cli.Context) map[string][]byte {
	rawSearchAttrKey := c.String(FlagSearchAttributesKey)
	var searchAttrKeys []string
	if strings.TrimSpace(rawSearchAttrKey) != "" {
		searchAttrKeys = trimSpace(strings.Split(rawSearchAttrKey, searchAttrInputSeparator))
	}

	rawSearchAttrVal := c.String(FlagSearchAttributesVal)
	var searchAttrVals []interface{}
	if strings.TrimSpace(rawSearchAttrVal) != "" {
		searchAttrValsStr := trimSpace(strings.Split(rawSearchAttrVal, searchAttrInputSeparator))

		for _, v := range searchAttrValsStr {
			searchAttrVals = append(searchAttrVals, convertStringToRealType(v))
		}
	}

	if len(searchAttrKeys) != len(searchAttrVals) {
		ErrorAndExit("Number of search attributes keys and values are not equal.", nil)
	}

	fields := map[string][]byte{}
	for i, key := range searchAttrKeys {
		val, err := json.Marshal(searchAttrVals[i])
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Encode value %v error", val), err)
		}
		fields[key] = val
	}

	return fields
}

func processMemo(c *cli.Context) map[string][]byte {
	rawMemoKey := c.String(FlagMemoKey)
	var memoKeys []string
	if strings.TrimSpace(rawMemoKey) != "" {
		memoKeys = strings.Split(rawMemoKey, " ")
	}

	rawMemoValue := processJSONInputHelper(c, jsonTypeMemo)
	var memoValues []string

	var sc fastjson.Scanner
	sc.Init(rawMemoValue)
	for sc.Next() {
		memoValues = append(memoValues, sc.Value().String())
	}
	if err := sc.Error(); err != nil {
		ErrorAndExit("Parse json error.", err)
	}
	if len(memoKeys) != len(memoValues) {
		ErrorAndExit("Number of memo keys and values are not equal.", nil)
	}

	fields := map[string][]byte{}
	for i, key := range memoKeys {
		fields[key] = []byte(memoValues[i])
	}
	return fields
}

func getPrintableMemo(memo *s.Memo) string {
	buf := new(bytes.Buffer)
	for k, v := range memo.Fields {
		fmt.Fprintf(buf, "%s=%s\n", k, string(v))
	}
	return buf.String()
}

func getPrintableSearchAttr(searchAttr *s.SearchAttributes) string {
	buf := new(bytes.Buffer)
	for k, v := range searchAttr.IndexedFields {
		var decodedVal interface{}
		json.Unmarshal(v, &decodedVal)
		fmt.Fprintf(buf, "%s=%v\n", k, decodedVal)
	}
	return buf.String()
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

func printRunStatus(event *s.HistoryEvent) {
	switch event.GetEventType() {
	case s.EventTypeWorkflowExecutionCompleted:
		fmt.Printf("  Status: %s\n", colorGreen("COMPLETED"))
		fmt.Printf("  Output: %s\n", string(event.WorkflowExecutionCompletedEventAttributes.Result))
	case s.EventTypeWorkflowExecutionFailed:
		fmt.Printf("  Status: %s\n", colorRed("FAILED"))
		fmt.Printf("  Reason: %s\n", event.WorkflowExecutionFailedEventAttributes.GetReason())
		fmt.Printf("  Detail: %s\n", string(event.WorkflowExecutionFailedEventAttributes.Details))
	case s.EventTypeWorkflowExecutionTimedOut:
		fmt.Printf("  Status: %s\n", colorRed("TIMEOUT"))
		fmt.Printf("  Timeout Type: %s\n", event.WorkflowExecutionTimedOutEventAttributes.GetTimeoutType())
	case s.EventTypeWorkflowExecutionCanceled:
		fmt.Printf("  Status: %s\n", colorRed("CANCELED"))
		fmt.Printf("  Detail: %s\n", string(event.WorkflowExecutionCanceledEventAttributes.Details))
	}
}

// in case workflow type is too long to show in table, trim it like .../example.Workflow
func trimWorkflowType(str string) string {
	res := str
	if len(str) >= maxWorkflowTypeLength {
		items := strings.Split(str, "/")
		res = items[len(items)-1]
		if len(res) >= maxWorkflowTypeLength {
			res = "..." + res[len(res)-maxWorkflowTypeLength:]
		} else {
			res = ".../" + res
		}
	}
	return res
}

func clustersToString(clusters []*shared.ClusterReplicationConfiguration) string {
	var res string
	for i, cluster := range clusters {
		if i == 0 {
			res = res + cluster.GetClusterName()
		} else {
			res = res + ", " + cluster.GetClusterName()
		}
	}
	return res
}

func getWorkflowStatus(statusStr string) s.WorkflowExecutionCloseStatus {
	if status, ok := workflowClosedStatusMap[strings.ToLower(statusStr)]; ok {
		return status
	}
	ErrorAndExit(optionErr, errors.New("option status is not one of allowed values "+
		"[completed, failed, canceled, terminated, continueasnew, timedout]"))
	return 0
}

func getWorkflowIDReusePolicy(value int) *s.WorkflowIdReusePolicy {
	if value >= 0 && value <= len(s.WorkflowIdReusePolicy_Values()) {
		return s.WorkflowIdReusePolicy(value).Ptr()
	}
	// At this point, the policy should return if the value is valid
	ErrorAndExit(fmt.Sprintf("Option %v value is not in supported range.", FlagWorkflowIDReusePolicy), nil)
	return nil
}

func archivalStatus(c *cli.Context) *shared.ArchivalStatus {
	if c.IsSet(FlagArchivalStatus) {
		switch c.String(FlagArchivalStatus) {
		case "disabled":
			return common.ArchivalStatusPtr(shared.ArchivalStatusDisabled)
		case "enabled":
			return common.ArchivalStatusPtr(shared.ArchivalStatusEnabled)
		default:
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagArchivalStatus), errors.New("invalid status, valid values are \"disabled\" and \"enabled\""))
		}
	}
	return nil
}
