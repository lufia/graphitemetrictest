// Package graphitepickletest provides utilities to test Graphite Pickle Protocol message.
package graphitepickletest

import (
	"time"
)

// Operator represents comparison operators.
type Operator uint8

// Operators.
const (
	Equal Operator = iota
	LessThan
	LessEqual
	GreaterThan
	GreaterEqual
)

// Expr represents a expression.
type Expr struct {
	Op    Operator
	Value float64
}

// Rule represents a rule for matching each lines in the protocol message.
type Rule struct {
	Required bool    // whether a rule should match to the message at least once.
	Path     string  // dot separated path; it can be contained some wildcards (* or #).
	Exprs    []*Expr // if Exprs is empty, that rule only checks the path exists.
}

// Metric represents a metric of the protocol.
type Metric struct {
	Path      string
	Timestamp time.Time
	Value     float64
}

// InvalidData contains invalid data.
//
// If Rule is not nil and Metric is not nil, the metric is violated for a rule's expression.
// If Rule is not nil and Metric is nil, the metric is needed but it is not found.
// If Rule is nil and Metric is not nil, the metric was not matched any rules.
type InvalidData struct {
	Rule   *Rule
	Metric *Metric
}

// Match checks validity of rules and metrics and returns any invalid data.
func Match(rules []*Rule, metrics []*Metric) ([]*InvalidData, error) {
	return nil, nil
}
