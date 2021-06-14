package graphitepickletest

import (
	"testing"
)

func TestMatch(t *testing.T) {
	r, err := Match([]*Rule{}, []*Metric{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 0 {
		t.Errorf("should match all rules if empty rules and metrics; but found invalid data: %v", r)
	}
}
