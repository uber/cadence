package log

// Fields is a K-V map for logging tags
type Fields map[tagType]interface{}

// Logger is our abstraction for logging
// Usage examples:
//  1) logger = logger.WithFields(log.Fields{
//		 log.TagWorkflowEventID: 123,
//		 log.TagDomainID       : "test-domain-id"})
//     logger.Info("hello world")
//	2) logger.Info("hello world", log.Fields{
//		log.TagWorkflowEventID : 123,
//		log.TagDomainID        : "test-domain-id",
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

	// We provide shortcuts for logging system lifecycle and major event
	// lifecycle
	DebugWithLifecycle(msg string, lifecycle string, tags ...Fields)
	InfoWithLifecycle(msg string, lifecycle string, tags ...Fields)
	WarnWithLifecycle(msg string, lifecycle string, tags ...Fields)
	ErrorWithLifecycle(msg string, lifecycle string, tags ...Fields)
	FatalWithLifecycle(msg string, lifecycle string, tags ...Fields)
	// major event
	DebugWithMajorEvent(msg string, majorEvent string, tags ...Fields)
	InfoWithMajorEvent(msg string, majorEvent string, tags ...Fields)
	WarnWithMajorEvent(msg string, majorEvent string, tags ...Fields)
	ErrorWithMajorEvent(msg string, majorEvent string, tags ...Fields)
	FatalWithMajorEvent(msg string, majorEvent string, tags ...Fields)
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
)

// helper function to define tags
func newTagType(name string, valType valueTypeEnum) tagType {
	return tagType{
		name:      name,
		valueType: valType,
	}
}
