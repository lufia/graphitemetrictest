package graphitemetrictest

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseError(t *testing.T) {
	e := &ParseError{
		Line: 1,
		Err:  errors.New("err"),
	}
	want := "parse error on line 1: err"
	if s := e.Error(); s != want {
		t.Errorf("Error() = %q; want %q", s, want)
	}
	if err := e.Unwrap(); err != e.Err {
		t.Errorf("Unwrap() = %v; want %v", err, e.Err)
	}
}

func TestReadMetrics(t *testing.T) {
	tests := []struct {
		in      string
		metrics []*Metric
	}{
		{
			in: "a.b.c 0 1623988183\n",
			metrics: []*Metric{
				{Path: "a.b.c", Value: 0.0, Timestamp: 1623988183},
			},
		},
		{
			in: "a.b.c 0 1623988183", // without '\n'
			metrics: []*Metric{
				{Path: "a.b.c", Value: 0.0, Timestamp: 1623988183},
			},
		},
		{
			in: "aa.bb.cc 0.0 -1\n",
			metrics: []*Metric{
				{Path: "aa.bb.cc", Value: 0.0, Timestamp: -1},
			},
		},
		{
			in: "a.b.c 0 1623988183\naa.bb.cc 0.0 -1\n",
			metrics: []*Metric{
				{Path: "a.b.c", Value: 0.0, Timestamp: 1623988183},
				{Path: "aa.bb.cc", Value: 0.0, Timestamp: -1},
			},
		},
		{
			in:      "\n",
			metrics: nil,
		},
	}
	for _, tt := range tests {
		f := strings.NewReader(tt.in)
		a, err := ReadMetrics(f)
		if err != nil {
			t.Fatalf("ReadMetrics(%q): %v", tt.in, err)
		}
		if !reflect.DeepEqual(a, tt.metrics) {
			t.Errorf("ReadMetrics(%q) = %v; want %v", tt.in, a, tt.metrics)
		}
	}
}

func TestReadMetrics_error(t *testing.T) {
	tests := []string{
		"a.b.c 0\n",              // 2 fields
		"a.b.c 0 1623988183 a\n", // 4 fields
		"a.b.c vvv 1623988183\n", // invalid value
		"a.b.c 0 aaa\n",          // invalid timestamp
	}
	for _, tt := range tests {
		f := strings.NewReader(tt)
		_, err := ReadMetrics(f)
		if err == nil {
			t.Fatalf("ReadMetrics(%q) should return an error ", tt)
		}
	}
}

func TestReadRules(t *testing.T) {
	tests := []struct {
		in    string
		rules []*Rule
	}{
		{
			in: "a.b.c >0",
			rules: []*Rule{
				{
					Required: true,
					Path:     "a.b.c",
					Exprs: []*Expr{
						{Op: GreaterThan, Value: 0.0},
					},
				},
			},
		},
		{
			in: "~a.b.c",
			rules: []*Rule{
				{
					Path: "a.b.c",
				},
			},
		},
		{
			in: "a.b.c\nd.e.f\n",
			rules: []*Rule{
				{
					Required: true,
					Path:     "a.b.c",
				},
				{
					Required: true,
					Path:     "d.e.f",
				},
			},
		},
		{
			in: "a.b.c <0.2, <=.3, >0,>=10.00",
			rules: []*Rule{
				{
					Required: true,
					Path:     "a.b.c",
					Exprs: []*Expr{
						{Op: LessThan, Value: 0.2},
						{Op: LessEqual, Value: 0.3},
						{Op: GreaterThan, Value: 0.0},
						{Op: GreaterEqual, Value: 10.0},
					},
				},
			},
		},
		{
			in: "//comment\na.b.c.xyz",
			rules: []*Rule{
				{
					Required: true,
					Path:     "a.b.c.xyz",
					Exprs:    nil,
				},
			},
		},
		{
			in: "a.b.c.xyz // comment",
			rules: []*Rule{
				{
					Required: true,
					Path:     "a.b.c.xyz",
					Exprs:    nil,
				},
			},
		},
	}
	for _, tt := range tests {
		f := strings.NewReader(tt.in)
		a, err := ReadRules(f)
		if err != nil {
			t.Fatalf("ReadRules(%q): %v", tt.in, err)
		}
		if !reflect.DeepEqual(a, tt.rules) {
			t.Errorf("ReadRules(%q) = %v; want %v", tt.in, a, tt.rules)
		}
	}
}
