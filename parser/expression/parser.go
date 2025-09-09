package expression

import (
	"errors"
	"fmt"
	"strings"
)

var ErrUnexpectedToken = errors.New("unexpected token in expression")

type parser struct {
	tokens []string
	pos    int
}

// Parse takes a raw filter string and returns an Expression AST
func Parse(input string) (Expression, error) {
	tokens := tokenize(input)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens to parse")
	}
	p := parser{tokens: tokens}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.hasMore() {
		return nil, fmt.Errorf("%w: %q", ErrUnexpectedToken, p.peek())
	}
	return expr, nil
}

// --- Helper methods ---

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

func (p *parser) hasMore() bool {
	return p.pos < len(p.tokens)
}

// --- Recursive descent parser ---

func (p *parser) parseExpression() (Expression, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		switch tok := strings.ToLower(p.peek()); tok {
		case "or", "||":
			p.next()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = Or{left, right}
		default:
			return left, nil
		}
	}
}

func (p *parser) parseTerm() (Expression, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		switch tok := strings.ToLower(p.peek()); tok {
		case "and", "&&":
			p.next()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			left = And{left, right}
		default:
			return left, nil
		}
	}
}

func (p *parser) parseFactor() (Expression, error) {
	tok := strings.ToLower(p.peek())
	switch tok {
	case "not", "!":
		p.next()
		e, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		return Not{e}, nil
	case "(":
		p.next()
		e, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.next() != ")" {
			return nil, fmt.Errorf("%w: expected ')'", ErrUnexpectedToken)
		}
		return e, nil
	default:
		return p.parseCondition()
	}
}

func (p *parser) parseCondition() (Expression, error) {
	if p.pos+2 >= len(p.tokens) {
		return nil, fmt.Errorf("%w: incomplete condition", ErrUnexpectedToken)
	}
	field := p.next()
	op := p.next()
	value := p.next()
	return Condition{
		Field: field,
		Op:    op,
		Value: value,
	}, nil
}
