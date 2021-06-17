package graphitemetrictest

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestRule_String(t *testing.T) {
	tests := []struct {
		name string
		rule *Rule
		s    string
	}{
		{
			name: "optional",
			rule: &Rule{
				Path: "a.b.c",
				Exprs: []*Expr{
					{Op: LessThan, Value: 3.0},
				},
			},
			s: "~a.b.c[<3]",
		},
		{
			name: "required",
			rule: &Rule{
				Required: true,
				Path:     "a.b.c",
				Exprs: []*Expr{
					{Op: LessThan, Value: 3.0},
				},
			},
			s: "a.b.c[<3]",
		},
		{
			name: "operators",
			rule: &Rule{
				Required: true,
				Path:     "a.b.c",
				Exprs: []*Expr{
					{Op: LessThan, Value: 3.0},
					{Op: LessEqual, Value: 2.15},
					{Op: GreaterThan, Value: 0.0},
					{Op: GreaterEqual, Value: -3.0},
				},
			},
			s: "a.b.c[<3,<=2.15,>0,>=-3]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.rule.String()
			if s != tt.s {
				t.Errorf("%+v = %q; want %q", tt.rule, s, tt.s)
			}
		})
	}
}

func TestMatch_expr(t *testing.T) {
	tests := []struct {
		exprs []*Expr
		value float64
		valid bool
	}{
		{
			exprs: []*Expr{
				{Op: LessThan, Value: 3.0},
			},
			value: 2.0,
			valid: true,
		},
		{
			exprs: []*Expr{
				{Op: LessThan, Value: 3.0},
			},
			value: 3.0,
			valid: false,
		},
		{
			exprs: []*Expr{
				{Op: LessEqual, Value: 3.0},
			},
			value: 3.0,
			valid: true,
		},
		{
			exprs: []*Expr{
				{Op: GreaterThan, Value: 3.0},
			},
			value: 3.1,
			valid: true,
		},
		{
			exprs: []*Expr{
				{Op: GreaterEqual, Value: 3.0},
			},
			value: 3.0,
			valid: true,
		},
	}
	for _, tt := range tests {
		rule := &Rule{
			Required: true,
			Path:     "a.b.c",
			Exprs:    tt.exprs,
		}
		metric := &Metric{
			Path:      "a.b.c",
			Value:     tt.value,
			Timestamp: time.Now().Unix(),
		}
		a := Match([]*Rule{rule}, []*Metric{metric})
		if tt.valid {
			if len(a) > 0 {
				t.Errorf("a value %g was not matched for rule={%v}", tt.value, rule)
			}
		} else {
			if len(a) == 0 {
				t.Errorf("a value %g was matched for rule={%v}; but it shouldn't", tt.value, rule)
			}
		}
	}
}

func TestMatch_path(t *testing.T) {
	tests := []struct {
		name    string
		rules   []*Rule
		metrics []*Metric
		want    []*InvalidData
	}{
		{
			name:    "empty",
			rules:   nil,
			metrics: nil,
			want:    nil,
		},
		{
			name: "simple path/success",
			rules: []*Rule{
				{
					Required: true,
					Path:     "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.value", Value: 3.0},
			},
			want: nil,
		},
		{
			name: "simple path/failure",
			rules: []*Rule{
				{
					Required: true,
					Path:     "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
				{
					Required: true,
					Path:     "custom.metric3.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 2.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.value1", Value: 3.0},
				{Path: "custom.metric2.value", Value: 3.0},
				{Path: "custom.metric1", Value: 3.0},
				{Path: "custom.metric3.value", Value: 3.0},
			},
			want: []*InvalidData{
				{
					Rule: &Rule{
						Required: true,
						Path:     "custom.metric1.value",
						Exprs: []*Expr{
							{Op: LessEqual, Value: 3.0},
						},
					},
				},
				{
					Metric: &Metric{Path: "custom.metric1.value1", Value: 3.0},
				},
				{
					Metric: &Metric{Path: "custom.metric2.value", Value: 3.0},
				},
				{
					Metric: &Metric{Path: "custom.metric1", Value: 3.0},
				},
				{
					Rule: &Rule{
						Required: true,
						Path:     "custom.metric3.value",
						Exprs: []*Expr{
							{Op: LessEqual, Value: 2.0},
						},
					},
					Metric: &Metric{Path: "custom.metric3.value", Value: 3.0},
				},
			},
		},
		{
			name: "optional path/matched",
			rules: []*Rule{
				{
					Path: "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.value", Value: 3.0},
			},
			want: []*InvalidData{},
		},
		{
			name: "optional path/passed",
			rules: []*Rule{
				{
					Path: "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: nil,
			want:    nil,
		},
		{
			name: "wildcard path",
			rules: []*Rule{
				{
					Required: true,
					Path:     "custom.#.writes.*",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
				{
					Required: true,
					Path:     "custom.#.reads.*",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.writes.bytes", Value: 3.0},
				{Path: "custom.metric2.writes.blocks", Value: 3.0},
			},
			want: []*InvalidData{
				{
					Rule: &Rule{
						Required: true,
						Path:     "custom.#.reads.*",
						Exprs: []*Expr{
							{Op: LessEqual, Value: 3.0},
						},
					},
				},
			},
		},
		{
			name: "same paths(OR condition)",
			rules: []*Rule{
				{
					Required: true,
					Path:     "custom.disk1.writes.bytes",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 0.0},
					},
				},
				{
					Required: true,
					Path:     "custom.disk1.writes.bytes",
					Exprs: []*Expr{
						{Op: GreaterEqual, Value: 512.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.disk1.writes.bytes", Value: 1024.0},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Match(tt.rules, tt.metrics)
			checkResults(t, "only result", a, tt.want)
			checkResults(t, "only expected", tt.want, a)
		})
	}
}

func checkResults(t *testing.T, name string, a1, a2 []*InvalidData) {
	t.Helper()
	if diffs := diffInvalidData(a1, a2); len(diffs) > 0 {
		t.Run(name, func(t *testing.T) {
			for _, d := range diffs {
				t.Errorf("unexpected %v", d)
			}
		})
	}
}

// diffInvalidData returns a slice in a part of the a1 that is *NOT* contained in a2.
func diffInvalidData(a1, a2 []*InvalidData) []*InvalidData {
	var x []*InvalidData
	for _, v1 := range a1 {
		var found bool
		for _, v2 := range a2 {
			if reflect.DeepEqual(v1, v2) {
				found = true
			}
		}
		if !found {
			x = append(x, v1)
		}
	}
	return x
}

func (p *InvalidData) String() string {
	s := ""
	if p.Rule != nil {
		s += fmt.Sprintf("rule = %v", p.Rule)
	}
	if p.Metric != nil {
		if s != "" {
			s += " "
		}
		s += fmt.Sprintf("value = %v", p.Metric)
	}
	return s
}
