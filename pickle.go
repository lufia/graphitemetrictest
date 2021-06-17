// Package graphitemetrictest provides utilities to test Graphite Plaintext Protocol message.
package graphitemetrictest

import (
	"fmt"
	"strings"
)

// The path accepts wildcards both '*' and '#'.
// But we always use '*' for internal key for path.
const (
	anyChar  = "*"
	anyChars = "*#"
)

// Operator represents comparison operators.
type Operator uint8

// Operators.
const (
	LessThan Operator = iota
	LessEqual
	GreaterThan
	GreaterEqual
)

// String returns the representation of the operator.
func (op Operator) String() string {
	switch op {
	case LessThan:
		return "<"
	case LessEqual:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterEqual:
		return ">="
	default:
		panic("unknown operator")
	}
}

// Expr represents a expression.
type Expr struct {
	Op    Operator
	Value float64
}

func (e *Expr) isValid(value float64) bool {
	switch e.Op {
	case LessThan:
		return value < e.Value
	case LessEqual:
		return value <= e.Value
	case GreaterThan:
		return value > e.Value
	case GreaterEqual:
		return value >= e.Value
	default:
		panic("unknown operator")
	}
}

// Rule represents a rule for matching each lines in the protocol message.
//
// BUGS(lufia): currently Path does not support tags syntax.
type Rule struct {
	Required bool    // whether a rule should match to the message at least once.
	Path     string  // dot separated path; it can be contained some wildcards (* or #).
	Exprs    []*Expr // if Exprs is empty, that rule only checks the path exists.
}

// String returns the string representation of the rule.
func (r *Rule) String() string {
	exprs := make([]string, len(r.Exprs))
	for i, e := range r.Exprs {
		exprs[i] = fmt.Sprintf("%s%g", e.Op, e.Value)
	}

	flag := ""
	if !r.Required {
		flag = "~"
	}
	return fmt.Sprintf("%s%s[%v]", flag, r.Path, strings.Join(exprs, ","))
}

// IsValid returns true if all expression are passed.
func (r *Rule) IsValid(value float64) bool {
	for _, e := range r.Exprs {
		if !e.isValid(value) {
			return false
		}
	}
	return true
}

// Metric represents a metric of the protocol.
type Metric struct {
	Path      string
	Value     float64
	Timestamp int64
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

type ruleMap struct {
	tree map[string]*ruleMap

	rules    []*Rule
	required bool
	used     int
}

func (m *ruleMap) isLeaf() bool {
	return m.tree == nil
}

func (m *ruleMap) leaves() []*ruleMap {
	var a []*ruleMap

	if m.isLeaf() {
		return nil
	}
	for _, v := range m.tree {
		if v.isLeaf() {
			a = append(a, v)
			continue
		}
		a = append(a, v.leaves()...)
	}
	return a
}

// isValid returns true if any one of the rules.
func (m *ruleMap) isValid(value float64) bool {
	for _, r := range m.rules {
		if r.IsValid(value) {
			return true
		}
	}
	return false
}

func (m *ruleMap) addRule(p []string, r *Rule) {
	for _, s := range p {
		if m.tree == nil {
			m.tree = make(map[string]*ruleMap)
		}
		v, ok := m.tree[s]
		if !ok {
			v = &ruleMap{}
			m.tree[s] = v
		}
		m = v
	}
	m.rules = append(m.rules, r)
	if r.Required {
		m.required = true
	}
}

func (m *ruleMap) lookupPath(p []string) *ruleMap {
	for _, s := range p {
		v, ok := m.tree[s]
		if !ok {
			v, ok = m.tree[anyChar]
			if !ok {
				return nil
			}
		}
		m = v
	}
	return m
}

func makeRules(rules []*Rule) *ruleMap {
	var m ruleMap
	for _, r := range rules {
		p := splitMetricName(r.Path)
		m.addRule(p, r)
	}
	return &m
}

func splitMetricName(s string) []string {
	a := strings.Split(s, ".")
	for i, v := range a {
		if strings.Contains(anyChars, v) {
			a[i] = anyChar
		}
	}
	return a
}

// Match checks validity of rules and metrics and returns any invalid data.
func Match(rules []*Rule, metrics []*Metric) []*InvalidData {
	var results []*InvalidData

	m := makeRules(rules)
	for _, c := range metrics {
		p := splitMetricName(c.Path)
		v := m.lookupPath(p)
		if v == nil || !v.isLeaf() {
			results = append(results, &InvalidData{Metric: c})
			continue
		}
		v.used++
		if !v.isValid(c.Value) {
			for _, r := range v.rules {
				results = append(results, &InvalidData{Rule: r, Metric: c})
			}
			continue
		}
	}
	for _, l := range m.leaves() {
		if l.required && l.used == 0 {
			for _, r := range l.rules {
				results = append(results, &InvalidData{Rule: r})
			}
		}
	}
	return results
}
