// Package graphitepickletest provides utilities to test Graphite Pickle Protocol message.
package graphitepickletest

import (
	"fmt"
	"strings"
	"time"
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
	return fmt.Sprintf("%s%s:%v", flag, r.Path, strings.Join(exprs, ","))
}

// Metric represents a metric of the protocol.
type Metric struct {
	Path      string
	Value     float64
	Timestamp time.Time
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
		r := m.lookupPath(p)
		if r == nil || !r.isLeaf() {
			results = append(results, &InvalidData{Metric: c})
			continue
		}
		r.used++
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
