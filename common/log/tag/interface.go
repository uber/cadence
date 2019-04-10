package tag

// Tag is the interface for logging system
type Tag interface {
	GetKey() string
	GetValue() interface{}
	GetValueType() valueTypeEnum
}

// the tag value types supported
const (
	// ValueTypeString is string value type
	ValueTypeString valueTypeEnum = 1
	// ValueTypeInteger is any value of the types(int8,int16,int32,unit8,uint32,uint64) will be converted into int64
	ValueTypeInteger valueTypeEnum = 2
	// ValueTypeDouble is for both float and double will be converted into double
	ValueTypeDouble valueTypeEnum = 3
	// ValueTypeBool is bool value type
	ValueTypeBool valueTypeEnum = 4
	// ValueTypeError is error type value
	ValueTypeError valueTypeEnum = 5
	// ValueTypeDuration is duration type value
	ValueTypeDuration valueTypeEnum = 6
	// ValueTypeTime is time type value
	ValueTypeTime valueTypeEnum = 7
	// ValueTypeObject will be converted into string by fmt.Sprintf("%+v")
	ValueTypeObject valueTypeEnum = 8

	// below are pre-defined value types(in values.go)
	ValueTypeWorkflowAction         valueTypeEnum = 9
	ValueTypeWorkflowListFilterType valueTypeEnum = 10
	ValueTypeSysComponent           valueTypeEnum = 11
	ValueTypeSysLifecycle           valueTypeEnum = 12
	ValueTypeSysErrorType           valueTypeEnum = 13
	ValueTypeSysShardUpdate         valueTypeEnum = 14
	ValueTypeSysOperationResult     valueTypeEnum = 15
	ValueTypeSysStoreOperation      valueTypeEnum = 16
)

// predefined tag values are using module private types, their values are public, defined in values.go
type valueTypeWorkflowAction string
type valueTypeWorkflowListFilterType string
type valueTypeSysComponent string
type valueTypeSysLifecycle string
type valueTypeSysErrorType string
type valueTypeSysShardUpdate string
type valueTypeSysOperationResult string
type valueTypeSysStoreOperation string

// keep this module private
type valueTypeEnum int

// keep this module private
type tagImpl struct {
	key       string
	value     interface{}
	valueType valueTypeEnum
}

func (t *tagImpl) GetKey() string {
	return t.key
}

func (t *tagImpl) GetValue() interface{} {
	return t.value
}

func (t *tagImpl) GetValueType() valueTypeEnum {
	return t.valueType
}

func newTag(key string, value interface{}, valueType valueTypeEnum) Tag {
	return &tagImpl{
		key:       key,
		valueType: valueType,
		value:     value,
	}
}
