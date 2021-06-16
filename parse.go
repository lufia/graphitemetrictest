package graphitepickletest

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Parse reads r and returns rules.
func Parse(r io.Reader) ([]*Rule, error) {
	var rules []*Rule

	f := bufio.NewReader(r)
	for {
		rule, err := parseRule(f)
		if err != nil {
			return nil, err
		}
		if rule == nil {
			break
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

type tokenKind int

const (
	tokenTilde tokenKind = iota
	tokenLessThan
	tokenLessEqual
	tokenGreaterThan
	tokenGreaterEqual
	tokenText
	tokenNumber
	tokenComma
	tokenNewline
)

type token struct {
	kind tokenKind
	text string
}

func parseRule(r *bufio.Reader) (*Rule, error) {
	rule := Rule{
		Required: true,
	}

	/*
	 * metric path
	 */
	var peak *token
	for { // skip blank lines
		t, err := readToken(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, nil
			}
			return nil, err
		}
		if t.kind != tokenNewline {
			peak = t
			break
		}
	}
	t := peak
	if t.kind == tokenTilde {
		rule.Required = false
		var err error
		t, err = readToken(r)
		if err != nil {
			return nil, fmt.Errorf("cannot read a path: %w", err)
		}
	}
	if t.kind != tokenText {
		return nil, fmt.Errorf("expected a path, but got %s", t.text)
	}
	rule.Path = t.text

	/*
	 * expressions
	 */
	for {
		t, err := readToken(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if t.kind == tokenNewline {
			break
		}
		t1, err := readToken(r)
		if err != nil {
			return nil, fmt.Errorf("cannot read a number: %w", err)
		}
		if t1.kind != tokenNumber {
			return nil, fmt.Errorf("expected a number, but got %s", t.text)
		}
		n, err := strconv.ParseFloat(t1.text, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to a number: %w", err)
		}
		switch t.kind {
		default:
			return nil, fmt.Errorf("expected a operator, but got %s", t.text)
		case tokenLessThan:
			rule.Exprs = append(rule.Exprs, &Expr{Op: LessThan, Value: n})
		case tokenLessEqual:
			rule.Exprs = append(rule.Exprs, &Expr{Op: LessEqual, Value: n})
		case tokenGreaterThan:
			rule.Exprs = append(rule.Exprs, &Expr{Op: GreaterThan, Value: n})
		case tokenGreaterEqual:
			rule.Exprs = append(rule.Exprs, &Expr{Op: GreaterEqual, Value: n})
		}

		/*
		 * comma or '\n'
		 */
		t, err = readToken(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if t.kind == tokenNewline {
			break
		}
		if t.kind != tokenComma {
			return nil, fmt.Errorf("expected ',', but got %s", t.text)
		}
	}
	return &rule, nil
}

func readToken(r *bufio.Reader) (*token, error) {
	if err := skipFunc(r, isSpace); err != nil {
		return nil, err
	}
	c, _, err := r.ReadRune()
	if err != nil {
		return nil, err
	}
	switch {
	case c == '\n':
		return &token{kind: tokenNewline, text: "\\n"}, nil
	case c == '/':
		c1, _, err := r.ReadRune()
		if err != nil {
			return nil, err
		}
		if c1 == '/' {
			if err := skipFunc(r, isComment); err != nil {
				return nil, err
			}
			return readToken(r)
		}
		return nil, errors.New("unexpected '/'")
	case c == '~':
		return &token{kind: tokenTilde, text: "~"}, nil
	case c == '<':
		c1, _, err := r.ReadRune()
		if err != nil {
			return nil, err
		}
		if c1 == '=' {
			return &token{kind: tokenLessEqual, text: "<="}, nil
		}
		if err := r.UnreadRune(); err != nil {
			return nil, err
		}
		return &token{kind: tokenLessThan, text: "<"}, nil
	case c == '>':
		c1, _, err := r.ReadRune()
		if err != nil {
			return nil, err
		}
		if c1 == '=' {
			return &token{kind: tokenGreaterEqual, text: ">="}, nil
		}
		if err := r.UnreadRune(); err != nil {
			return nil, err
		}
		return &token{kind: tokenGreaterThan, text: ">"}, nil
	case c == ',':
		return &token{kind: tokenComma, text: ","}, nil
	case isNumber(c):
		if err := r.UnreadRune(); err != nil {
			return nil, err
		}
		return readText(r, isNumber, tokenNumber)
	default:
		if err := r.UnreadRune(); err != nil {
			return nil, err
		}
		return readText(r, isText, tokenText)
	}
}

func readText(r *bufio.Reader, f func(c rune) bool, kind tokenKind) (*token, error) {
	var w strings.Builder
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) && w.Len() > 0 {
				return &token{kind: kind, text: w.String()}, nil
			}
			return nil, err
		}
		if !f(c) {
			break
		}
		if _, err := w.WriteRune(c); err != nil {
			return nil, err
		}
	}
	if err := r.UnreadRune(); err != nil {
		return nil, err
	}
	return &token{kind: kind, text: w.String()}, nil
}

func isText(c rune) bool {
	return !unicode.IsSpace(c)
}

func isNumber(c rune) bool {
	return (c >= '0' && c <= '9') || c == '.'
}

func isComment(c rune) bool {
	return c != '\n'
}

func isSpace(c rune) bool {
	return unicode.IsSpace(c) && c != '\n'
}

func skipFunc(r *bufio.Reader, f func(c rune) bool) error {
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if !f(c) {
			r.UnreadRune()
			break
		}
	}
	return nil
}
