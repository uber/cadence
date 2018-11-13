---
layout: doc
itle: Testing workflows
weight: 27
---

# Test your workflow

The Cadence Go client library provides a test framework to facilitate testing workflow implementations.
The framework is suited for implementing unit tests as well as functional tests of the workflow logic.

The following code implements unit tests for the `SimpleWorkflow` sample:

```go
package sample

import (
        "errors"
        "testing"

        "code.uber.internal/devexp/cadence-worker/activity"

        "github.com/stretchr/testify/mock"
        "github.com/stretchr/testify/suite"

        "go.uber.org/cadence"
)

type UnitTestSuite struct {
        suite.Suite
        cadence.WorkflowTestSuite

        env *cadence.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
        s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
        s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_SimpleWorkflow_Success() {
        s.env.ExecuteWorkflow(SimpleWorkflow, "test_success")

        s.True(s.env.IsWorkflowCompleted())
        s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_SimpleWorkflow_ActivityParamCorrect() {
        s.env.OnActivity(SimpleActivity, mock.Anything, mock.Anything).Return(func(ctx context.Context, value string) (string, error) {
                s.Equal("test_success", value)
                return value, nil
        })
        s.env.ExecuteWorkflow(SimpleWorkflow, "test_success")

        s.True(s.env.IsWorkflowCompleted())
        s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_SimpleWorkflow_ActivityFails() {
        s.env.OnActivity(SimpleActivity, mock.Anything, mock.Anything).Return("", errors.New("SimpleActivityFailure"))
        s.env.ExecuteWorkflow(SimpleWorkflow, "test_failure")

        s.True(s.env.IsWorkflowCompleted())

        s.NotNil(s.env.GetWorkflowError())
        _, ok := s.env.GetWorkflowError().(*cadence.GenericError)
        s.True(ok)
        s.Equal("SimpleActivityFailure", s.env.GetWorkflowError().Error())
}

func TestUnitTestSuite(t *testing.T) {
        suite.Run(t, new(UnitTestSuite))
}
```

## Setup

To run unit tests, we first define a "test suite" struct that absorbs both the
basic suite functionality from [testify](https://godoc.org/github.com/stretchr/testify/suite)
via `suite.Suite` and the suite functionality from the Cadence test framework via
`cadence.WorkflowTestSuite`. Because every test in this test suite will test our workflow, we
add a property to our struct to hold an instance of the test environment. This allows us to initialize
the test environment in a setup method. For testing workflows, we use a `cadence.TestWorkflowEnvironment`.

Next, we implement a `SetupTest` method to setup a new test environment before each test. Doing so
ensures that each test runs in its own isolated sandbox. We also implement an `AfterTest` function
where we assert that all mocks we set up were indeed called by invoking `s.env.AssertExpectations(s.T()).

Finally, we create a regular test function recognized by "go test" and pass the struct to `suite.Run`.

## A simple test

The most simple test case we can write is to have the test environment execute the workflow and then
evaluate the results.

```go
func (s *UnitTestSuite) Test_SimpleWorkflow_Success() {
        s.env.ExecuteWorkflow(SimpleWorkflow, "test_success")

        s.True(s.env.IsWorkflowCompleted())
        s.NoError(s.env.GetWorkflowError())
}
```
Calling `s.env.ExecuteWorkflow(...)` executes the workflow logic and any invoked activities inside the
test process. The first parameter of `s.env.ExecuteWorkflow(...)` contains the workfflow functions,
and any subsequent parameters contain values for custom input parameters declared by the workflow
function.

<p class ="callout info">Note that unless the activity invocations are mocked or activity implementation
replaced (see [Activity mocking and overriding](#Activity-mocking-and-overriding)), the test environment
will execute the actual activity code including any calls to outside services.</p>

After executing the workflow in the above example, we assert that the workflow ran through completion
via the call to `s.env.IsWorkflowComplete()`. We also asser that no errors were returned by asserting
on the return value of `s.env.GetWorkflowError()`. If our workflow returned a value, we could have
retrieved that value via a call to `s.env.GetWorkflowResult(&value)` and had additional asserts on that
value.

## Activity mocking and overriding

When running unit tests on workflows, we want to test the workflow logic in isolation. Additionially,
we want to inject activity errors during our test runs. The test framework provides two mechanisms
that support these scenarios: activity mocking and activity overriding. Both of these mechanisms allow
you to change the behavior of activities invoked by your workflow without the need to modify the actual
workflow code.

Let's take a look at a test that simulates a test that fails via the "activity mocking" mechanism.

```go
func (s *UnitTestSuite) Test_SimpleWorkflow_ActivityFails() {
        s.env.OnActivity(SimpleActivity, mock.Anything, mock.Anything).Return("", errors.New("SimpleActivityFailure"))
        s.env.ExecuteWorkflow(SimpleWorkflow, "test_failure")

        s.True(s.env.IsWorkflowCompleted())

        s.NotNil(s.env.GetWorkflowError())
        _, ok := s.env.GetWorkflowError().(*cadence.GenericError)
        s.True(ok)
        s.Equal("SimpleActivityFailure", s.env.GetWorkflowError().Error())
}
```
This test simulates the execution of the activity `SimpleActivity` that is invoked by our workflow
`SimpleWorkflow` returning an error. We accomplish this by setting up a mock on the test environment
for the `SimpleActivity` that returns an error.

```go
s.env.OnActivity(SimpleActivity, mock.Anything, mock.Anything).Return("", errors.New("SimpleActivityFailure"))
```
With the mock set up we can now execute the workflow via the s.env.ExecuteWorkflow(...) method and
assert that the workflow completed successfully and returned the expected error.

Simply mocking the execution to return a desired value or error is a pretty powerful mechnism to
isolate workflow logic. However, sometimes we want to replace the activity with an alternate implementation
to support a more complex test scenario. Let's assume we want to validate that the activity gets called
with the correct parameters.

```go
func (s *UnitTestSuite) Test_SimpleWorkflow_ActivityParamCorrect() {
        s.env.OnActivity(SimpleActivity, mock.Anything, mock.Anything).Return(func(ctx context.Context, value string) (string, error) {
                s.Equal("test_success", value)
                return value, nil
        })
        s.env.ExecuteWorkflow(SimpleWorkflow, "test_success")

        s.True(s.env.IsWorkflowCompleted())
        s.NoError(s.env.GetWorkflowError())
}
```

In this example, we provide a function implementation as the parameter to `Return`. This allows us to
provide an alternate implementation for the activity `SimpleActivity`. The framework will execute this
function whenever the activity is invoked and pass on the return value from the function as the result
of the activity invocation. Additionally, the framework will validate that the signature of the “mock”
function matches the signature of the original activity function.

Since this can be an entire function, there is no limitation as to what we can do in here. In this
example to assert that the “value” param has the same content to the value param we passed to the workflow.
