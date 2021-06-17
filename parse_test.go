package graphitemetrictest

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse_line(t *testing.T) {
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
		a, err := Parse(f)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.in, err)
		}
		if !reflect.DeepEqual(a, tt.rules) {
			t.Errorf("Parse(%q) = %v; want %v", tt.in, a, tt.rules)
		}
	}
}
