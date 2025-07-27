// Package parser implements a lightweight logical expression parser and evaluator.
// It supports conditions (==, !=, >, >=, <, <=, contains) with AND, OR, NOT logic.
//
// Example expression:
//
//	name = "John" AND age >= 21 OR NOT (email contains "@test.com")
//
// This can be evaluated against CSV/record fields for filtering email recipients.
package parser

import (
	"strconv"
	"strings"
)

// EXPRESSION is the interface that all logical expressions implement.
type EXPRESSION interface {
	Evaluate(map[string]string) bool
}

// The CONDITION represents a basic field comparison.
type CONDITION struct {
	Field string
	Op    string
	Value string
}

type AND struct{ LEFT, RIGHT EXPRESSION }
type OR struct{ LEFT, RIGHT EXPRESSION }
type NOT struct{ INNER EXPRESSION }

// Evaluate evaluates a CONDITION against the data map.
func (c CONDITION) Evaluate(data map[string]string) bool {
	val := data[c.Field]

	switch c.Op {
	case "=", "==":
		return val == c.Value
	case "!=":
		return val != c.Value
	case "contains":
		return strings.Contains(val, c.Value)
	case ">", ">=", "<", "<=":
		f1, err1 := strconv.ParseFloat(val, 64)
		f2, err2 := strconv.ParseFloat(c.Value, 64) // 🔧 Fixed here
		if err1 != nil || err2 != nil {
			return false
		}
		switch c.Op {
		case ">":
			return f1 > f2
		case ">=":
			return f1 >= f2
		case "<":
			return f1 < f2
		case "<=":
			return f1 <= f2
		}
	}
	return false
}

// Evaluate implements logical AND
func (a AND) Evaluate(data map[string]string) bool {
	return a.LEFT.Evaluate(data) && a.RIGHT.Evaluate(data)
}

// Evaluate implements logical OR
func (o OR) Evaluate(data map[string]string) bool {
	return o.LEFT.Evaluate(data) || o.RIGHT.Evaluate(data)
}

// Evaluate implements logical NOT
func (n NOT) Evaluate(data map[string]string) bool {
	return !n.INNER.Evaluate(data)
}

// Parse parses a space-separated logical expression string into an EXPRESSION tree.
//
// Grammar:
//
//	expr := term { OR term }
//	term: = factor { AND factor }
//	factor: = NOT factor | '(' expr ')' | cond
//	cond: = field op value
//
// Operators: = == != > >= < <= contains
func Parse(input string) (EXPRESSION, error) {
	tokens := strings.Fields(input)
	p := parser{tokens: tokens}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.pos != len(tokens) {
		return nil, ErrUnexpectedToken
	}
	return expr, nil
}

var ErrUnexpectedToken = strconv.ErrSyntax

type parser struct {
	tokens []string
	pos    int
}

func (p *parser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *parser) next() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	p.pos++
	return p.tokens[p.pos-1]
}

func (p *parser) parseExpression() (EXPRESSION, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for {
		tok := strings.ToUpper(p.peek())
		if tok == "OR" || tok == "||" {
			p.next()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = OR{left, right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseTerm() (EXPRESSION, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		tok := strings.ToUpper(p.peek())
		if tok == "AND" || tok == "&&" {
			p.next()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			left = AND{left, right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseFactor() (EXPRESSION, error) {
	tok := strings.ToUpper(p.peek())
	if tok == "NOT" || tok == "!" {
		p.next()
		e, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		return NOT{e}, nil
	}
	if tok == "(" {
		p.next()
		e, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.next() != ")" {
			return nil, ErrUnexpectedToken
		}
		return e, nil
	}
	return p.parseCond()
}

func (p *parser) parseCond() (EXPRESSION, error) {
	if p.pos+2 >= len(p.tokens) {
		return nil, ErrUnexpectedToken
	}
	field := p.next()
	op := p.next()
	value := p.next()
	return CONDITION{
		Field: field,
		Op:    op,
		Value: strings.Trim(value, `"`), // Trim surrounding quotes if present
	}, nil
}

// Filter filters recipients using the logical expression.
func Filter(recipients []Recipient, exp EXPRESSION) []Recipient {
	var out []Recipient
	for _, r := range recipients {
		data := map[string]string{"email": r.Email}
		for k, v := range r.Data {
			data[k] = v
		}
		if exp.Evaluate(data) {
			out = append(out, r)
		}
	}
	return out
}
