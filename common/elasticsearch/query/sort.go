package query

type Sorter interface {
	Source() (interface{}, error)
}

// FieldSort sorts by a given field.
type FieldSort struct {
	Sorter
	fieldName string
	ascending bool
}

// NewFieldSort creates a new FieldSort.
func NewFieldSort(fieldName string) *FieldSort {
	return &FieldSort{
		fieldName: fieldName,
		ascending: true,
	}
}

// FieldName specifies the name of the field to be used for sorting.
func (s *FieldSort) FieldName(fieldName string) *FieldSort {
	s.fieldName = fieldName
	return s
}

// Order defines whether sorting ascending (default) or descending.
func (s *FieldSort) Order(ascending bool) *FieldSort {
	s.ascending = ascending
	return s
}

// Asc sets ascending sort order.
func (s *FieldSort) Asc() *FieldSort {
	s.ascending = true
	return s
}

// Desc sets descending sort order.
func (s *FieldSort) Desc() *FieldSort {
	s.ascending = false
	return s
}

// Source returns the JSON-serializable data.
func (s *FieldSort) Source() (interface{}, error) {
	source := make(map[string]interface{})
	x := make(map[string]interface{})
	source[s.fieldName] = x
	if s.ascending {
		x["order"] = "asc"
	} else {
		x["order"] = "desc"
	}

	return source, nil
}
