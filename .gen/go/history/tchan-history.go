// @generated Code generated by thrift-gen. Do not modify.

// Package history is generated code used to make or handle TChannel calls using Thrift.
package history

import (
	"fmt"

	athrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/uber/tchannel-go/thrift"

	"github.com/uber/cadence/.gen/go/shared"
)

var _ = shared.GoUnusedProtection__

// Interfaces for the service and client for the services defined in the IDL.

// TChanHistoryService is the interface that defines the server handler and client interface.
type TChanHistoryService interface {
	GetWorkflowExecutionHistory(ctx thrift.Context, getRequest *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error)
	RecordActivityTaskHeartbeat(ctx thrift.Context, heartbeatRequest *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error)
	RecordActivityTaskStarted(ctx thrift.Context, addRequest *RecordActivityTaskStartedRequest) (*RecordActivityTaskStartedResponse, error)
	RecordDecisionTaskStarted(ctx thrift.Context, addRequest *RecordDecisionTaskStartedRequest) (*RecordDecisionTaskStartedResponse, error)
	RespondActivityTaskCanceled(ctx thrift.Context, canceledRequest *shared.RespondActivityTaskCanceledRequest) error
	RespondActivityTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondActivityTaskCompletedRequest) error
	RespondActivityTaskFailed(ctx thrift.Context, failRequest *shared.RespondActivityTaskFailedRequest) error
	RespondDecisionTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondDecisionTaskCompletedRequest) error
	StartWorkflowExecution(ctx thrift.Context, startRequest *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error)
}

// Implementation of a client and service handler.

type tchanHistoryServiceClient struct {
	thriftService string
	client        thrift.TChanClient
}

func NewTChanHistoryServiceInheritedClient(thriftService string, client thrift.TChanClient) *tchanHistoryServiceClient {
	return &tchanHistoryServiceClient{
		thriftService,
		client,
	}
}

// NewTChanHistoryServiceClient creates a client that can be used to make remote calls.
func NewTChanHistoryServiceClient(client thrift.TChanClient) TChanHistoryService {
	return NewTChanHistoryServiceInheritedClient("HistoryService", client)
}

func (c *tchanHistoryServiceClient) GetWorkflowExecutionHistory(ctx thrift.Context, getRequest *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	var resp HistoryServiceGetWorkflowExecutionHistoryResult
	args := HistoryServiceGetWorkflowExecutionHistoryArgs{
		GetRequest: getRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "GetWorkflowExecutionHistory", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanHistoryServiceClient) RecordActivityTaskHeartbeat(ctx thrift.Context, heartbeatRequest *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	var resp HistoryServiceRecordActivityTaskHeartbeatResult
	args := HistoryServiceRecordActivityTaskHeartbeatArgs{
		HeartbeatRequest: heartbeatRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RecordActivityTaskHeartbeat", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanHistoryServiceClient) RecordActivityTaskStarted(ctx thrift.Context, addRequest *RecordActivityTaskStartedRequest) (*RecordActivityTaskStartedResponse, error) {
	var resp HistoryServiceRecordActivityTaskStartedResult
	args := HistoryServiceRecordActivityTaskStartedArgs{
		AddRequest: addRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RecordActivityTaskStarted", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EventAlreadyStartedError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanHistoryServiceClient) RecordDecisionTaskStarted(ctx thrift.Context, addRequest *RecordDecisionTaskStartedRequest) (*RecordDecisionTaskStartedResponse, error) {
	var resp HistoryServiceRecordDecisionTaskStartedResult
	args := HistoryServiceRecordDecisionTaskStartedArgs{
		AddRequest: addRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RecordDecisionTaskStarted", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EventAlreadyStartedError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanHistoryServiceClient) RespondActivityTaskCanceled(ctx thrift.Context, canceledRequest *shared.RespondActivityTaskCanceledRequest) error {
	var resp HistoryServiceRespondActivityTaskCanceledResult
	args := HistoryServiceRespondActivityTaskCanceledArgs{
		CanceledRequest: canceledRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondActivityTaskCanceled", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanHistoryServiceClient) RespondActivityTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondActivityTaskCompletedRequest) error {
	var resp HistoryServiceRespondActivityTaskCompletedResult
	args := HistoryServiceRespondActivityTaskCompletedArgs{
		CompleteRequest: completeRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondActivityTaskCompleted", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanHistoryServiceClient) RespondActivityTaskFailed(ctx thrift.Context, failRequest *shared.RespondActivityTaskFailedRequest) error {
	var resp HistoryServiceRespondActivityTaskFailedResult
	args := HistoryServiceRespondActivityTaskFailedArgs{
		FailRequest: failRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondActivityTaskFailed", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanHistoryServiceClient) RespondDecisionTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondDecisionTaskCompletedRequest) error {
	var resp HistoryServiceRespondDecisionTaskCompletedResult
	args := HistoryServiceRespondDecisionTaskCompletedArgs{
		CompleteRequest: completeRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondDecisionTaskCompleted", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanHistoryServiceClient) StartWorkflowExecution(ctx thrift.Context, startRequest *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	var resp HistoryServiceStartWorkflowExecutionResult
	args := HistoryServiceStartWorkflowExecutionArgs{
		StartRequest: startRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "StartWorkflowExecution", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.SessionAlreadyExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

type tchanHistoryServiceServer struct {
	handler TChanHistoryService
}

// NewTChanHistoryServiceServer wraps a handler for TChanHistoryService so it can be
// registered with a thrift.Server.
func NewTChanHistoryServiceServer(handler TChanHistoryService) thrift.TChanServer {
	return &tchanHistoryServiceServer{
		handler,
	}
}

func (s *tchanHistoryServiceServer) Service() string {
	return "HistoryService"
}

func (s *tchanHistoryServiceServer) Methods() []string {
	return []string{
		"GetWorkflowExecutionHistory",
		"RecordActivityTaskHeartbeat",
		"RecordActivityTaskStarted",
		"RecordDecisionTaskStarted",
		"RespondActivityTaskCanceled",
		"RespondActivityTaskCompleted",
		"RespondActivityTaskFailed",
		"RespondDecisionTaskCompleted",
		"StartWorkflowExecution",
	}
}

func (s *tchanHistoryServiceServer) Handle(ctx thrift.Context, methodName string, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	switch methodName {
	case "GetWorkflowExecutionHistory":
		return s.handleGetWorkflowExecutionHistory(ctx, protocol)
	case "RecordActivityTaskHeartbeat":
		return s.handleRecordActivityTaskHeartbeat(ctx, protocol)
	case "RecordActivityTaskStarted":
		return s.handleRecordActivityTaskStarted(ctx, protocol)
	case "RecordDecisionTaskStarted":
		return s.handleRecordDecisionTaskStarted(ctx, protocol)
	case "RespondActivityTaskCanceled":
		return s.handleRespondActivityTaskCanceled(ctx, protocol)
	case "RespondActivityTaskCompleted":
		return s.handleRespondActivityTaskCompleted(ctx, protocol)
	case "RespondActivityTaskFailed":
		return s.handleRespondActivityTaskFailed(ctx, protocol)
	case "RespondDecisionTaskCompleted":
		return s.handleRespondDecisionTaskCompleted(ctx, protocol)
	case "StartWorkflowExecution":
		return s.handleStartWorkflowExecution(ctx, protocol)

	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanHistoryServiceServer) handleGetWorkflowExecutionHistory(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceGetWorkflowExecutionHistoryArgs
	var res HistoryServiceGetWorkflowExecutionHistoryResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.GetWorkflowExecutionHistory(ctx, req.GetRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRecordActivityTaskHeartbeat(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRecordActivityTaskHeartbeatArgs
	var res HistoryServiceRecordActivityTaskHeartbeatResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.RecordActivityTaskHeartbeat(ctx, req.HeartbeatRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRecordActivityTaskStarted(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRecordActivityTaskStartedArgs
	var res HistoryServiceRecordActivityTaskStartedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.RecordActivityTaskStarted(ctx, req.AddRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *EventAlreadyStartedError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for eventAlreadyStartedError returned non-nil error type *EventAlreadyStartedError but nil value")
			}
			res.EventAlreadyStartedError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRecordDecisionTaskStarted(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRecordDecisionTaskStartedArgs
	var res HistoryServiceRecordDecisionTaskStartedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.RecordDecisionTaskStarted(ctx, req.AddRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *EventAlreadyStartedError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for eventAlreadyStartedError returned non-nil error type *EventAlreadyStartedError but nil value")
			}
			res.EventAlreadyStartedError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRespondActivityTaskCanceled(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRespondActivityTaskCanceledArgs
	var res HistoryServiceRespondActivityTaskCanceledResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondActivityTaskCanceled(ctx, req.CanceledRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRespondActivityTaskCompleted(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRespondActivityTaskCompletedArgs
	var res HistoryServiceRespondActivityTaskCompletedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondActivityTaskCompleted(ctx, req.CompleteRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRespondActivityTaskFailed(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRespondActivityTaskFailedArgs
	var res HistoryServiceRespondActivityTaskFailedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondActivityTaskFailed(ctx, req.FailRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleRespondDecisionTaskCompleted(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceRespondDecisionTaskCompletedArgs
	var res HistoryServiceRespondDecisionTaskCompletedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondDecisionTaskCompleted(ctx, req.CompleteRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanHistoryServiceServer) handleStartWorkflowExecution(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HistoryServiceStartWorkflowExecutionArgs
	var res HistoryServiceStartWorkflowExecutionResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.StartWorkflowExecution(ctx, req.StartRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.WorkflowExecutionAlreadyStartedError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for sessionAlreadyExistError returned non-nil error type *shared.WorkflowExecutionAlreadyStartedError but nil value")
			}
			res.SessionAlreadyExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}
