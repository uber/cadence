package log

// Fields is a K-V map for logging tags
type Fields map[tagType]interface{}

// Logger is our abstraction for logging
// Usage examples:
//  1) logger = logger.WithFields(log.Fields{
//          log.TagWorkflowEventID: 123,
//          log.TagDomainID       : "test-domain-id"})
//     logger.Info("hello world")
//  2) logger.Info("hello world", log.Fields{
//          log.TagWorkflowEventID : 123,
//          log.TagDomainID        : "test-domain-id",
//	   })
//  Note: msg should be static, it is not recommended to use fmt.Sprintf() for msg.
//        Anything dynamic should be tagged.
type Logger interface {
	Debug(msg string, tags ...Fields)
	Info(msg string, tags ...Fields)
	Warn(msg string, tags ...Fields)
	Error(msg string, tags ...Fields)
	Fatal(msg string, tags ...Fields)
	WithFields(tags Fields) Logger

	// We provide shortcuts for logging component lifecycle logs
	SetComponent(component tagValueTypeSysComponent) Logger
	DebugWithLifecycle(msg string, lifecycle tagValueTypeSysLifecycle, tags ...Fields)
	InfoWithLifecycle(msg string, lifecycle tagValueTypeSysLifecycle, tags ...Fields)
	WarnWithLifecycle(msg string, lifecycle tagValueTypeSysLifecycle, tags ...Fields)
	ErrorWithLifecycle(msg string, lifecycle tagValueTypeSysLifecycle, tags ...Fields)
	FatalWithLifecycle(msg string, lifecycle tagValueTypeSysLifecycle, tags ...Fields)
}

// we intentionally make tagType module-private to enforce using pre-defined tag keys
type tagType struct {
	name      string
	valueType valueTypeEnum
}

type valueTypeEnum string

// the tag value types supported
const (
	// string value type
	valueTypeString valueTypeEnum = "string"
	// any value of the types(int8,int16,int32,unit8,uint32,uint64) will be converted into int64
	valueTypeInteger valueTypeEnum = "integer"
	// both float and double will be converted into double
	valueTypeDouble valueTypeEnum = "double"
	// bool value type
	valueTypeBool valueTypeEnum = "bool"
	// error type value
	valueTypeError valueTypeEnum = "error"
	// duration type value
	valueTypeDuration valueTypeEnum = "duration"
	// time type value
	valueTypeTime valueTypeEnum = "time"
	// it will be converted into string by fmt.Sprintf("%+v")
	valueTypeObject valueTypeEnum = "object"

	// pre-defined value types(in tagValues.go)
	valueTypeWorkflowAction         valueTypeEnum = "value_for_workflow_action"
	valueTypeWorkflowListFilterType valueTypeEnum = "value_for_workflow_list_filter_type"
	// pre-defined value types(in tagValues.go)
	valueTypeSysComponent       valueTypeEnum = "value_for_sys_component"
	valueTypeSysLifecycle       valueTypeEnum = "value_for_sys_lifecycle"
	valueTypeSysErrorType       valueTypeEnum = "value_for_sys_error_type"
	valueTypeSysShardUpdate     valueTypeEnum = "value_for_sys_shard_update"
	valueTypeSysOperationResult valueTypeEnum = "value_for_sys_operation_result"
	valueTypeSysStoreOperation  valueTypeEnum = "value_for_sys_store_operation"
)

// helper function to define tags
func newTagType(name string, valType valueTypeEnum) tagType {
	return tagType{
		name:      name,
		valueType: valType,
	}
}

// predefined tag values are using module private types, their values are public, defined in tagValues.go
type tagValueTypeWorkflowAction string
type tagValueTypeWorkflowListFilterType string
type tagValueTypeSysComponent string
type tagValueTypeSysLifecycle string
type tagValueTypeSysErrorType string
type tagValueTypeSysShardUpdate string
type tagValueTypeOperationResult string
type tagValueTypeSysStoreOperation string
