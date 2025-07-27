package expression

import (
	"strconv"
	"strings"
)

type Expression interface {
	Evaluate(data map[string]string) bool
}

type Condition struct {
	Field string
	Op    string
	Value string
}

type And struct {
	Left, Right Expression
}

type Or struct {
	Left, Right Expression
}

type Not struct {
	Inner Expression
}

func (c Condition) Evaluate(data map[string]string) bool {
	val := strings.ToLower(data[strings.ToLower(c.Field)])
	cmp := strings.ToLower(c.Value)

	switch strings.ToLower(c.Op) {
	case "=", "==":
		if cmp == ".." {
			return val == ""
		}
		return val == cmp
	case "!=":
		if cmp == ".." {
			return val != ""
		}
		return val != cmp
	case "contains":
		return strings.Contains(val, cmp)
	case "startswith":
		return strings.HasPrefix(val, cmp)
	case "endswith":
		return strings.HasSuffix(val, cmp)
	case ">", ">=", "<", "<=":
		f1, err1 := strconv.ParseFloat(val, 64)
		f2, err2 := strconv.ParseFloat(c.Value, 64)
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

func (a And) Evaluate(data map[string]string) bool {
	return a.Left.Evaluate(data) && a.Right.Evaluate(data)
}

func (o Or) Evaluate(data map[string]string) bool {
	return o.Left.Evaluate(data) || o.Right.Evaluate(data)
}

func (n Not) Evaluate(data map[string]string) bool {
	return !n.Inner.Evaluate(data)
}
