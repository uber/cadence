package query

import (
	"encoding/json"
	"testing"
)

func TestBoolQuery(t *testing.T) {
	q := NewBoolQuery()
	q = q.MustNot(NewRangeQuery("age").From(10).To(20))
	src, err := q.Source()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"bool":{"must_not":[{"range":{"age":{"from":10,"include_lower":true,"include_upper":true,"to":20}}}]}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}
