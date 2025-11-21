package workflow

import (
	"fmt"
	"strings"
)

// Result holds the evaluation result and trace.
type Result struct {
	Value any
	Trace string
}

// Evaluator evaulates GitHub Actions expressions.
type Evaluator struct {
	ctx *Context
}

func NewEvaluator(ctx *Context) *Evaluator {
	return &Evaluator{
		ctx: ctx,
	}
}

func (e *Evaluator) Evaluate(expr string) (*Result, error) {
	// Strip ${{ }} wrapper if present.
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "${{") && strings.HasSuffix(expr, "}}") {
		expr = strings.TrimPrefix(expr, "${{")
		expr = strings.TrimSuffix(expr, "}}")
		expr = strings.TrimSpace(expr)
	}

	tokens, err := tokenize(expr)
	if err != nil {
		return nil, fmt.Errorf("tokenize: %w", err)
	}

	p := &parser{tokens: tokens}

	node, err := p.parse()
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}

	return e.eval(node)
}

func (e *Evaluator) eval(node Node) (*Result, error) {
	switch n := node.(type) {
	case *LiteralNode:
		return &Result{Value: n.Value, Trace: formatValue(n.Value)}, nil

	case *ContextAccessNode:
		val, ok := e.ctx.Lookup(n.Path)
		if !ok {
			return &Result{Value: nil, Trace: fmt.Sprintf("%s -> null", n.Path)}
		}
		return &Result{Value: val, Trace: fmt.Sprintf("%s -> %s", n.Path, formatValue(val))}, nil

	case *BinaryOpNode:
		left, err := e.eval(n.Left)
		if err != nil {
			return nil, err
		}

		right, err := e.eval(n.Right)
		if err != nil {
			return nil, err
		}

		result := applyBinaryOp(n.Op, left.Value, right.Value)
		trace := fmt.Sprintf("%s %s %s -> %s", left.Trace, n.Op, right.Trace, formatValue(result))
		return &Result{Value: result, Trace: trace}, nil

	case *FunctionalCallNode:
		var args []any
		var argTraces []string

		for _, arg := range n.Args {
			r, err := e.eval(arg)
			if err != nil {
				return nil, err
			}

			args = append(args, r.Value)
			argTraces = append(argTraces, r.Trace)
		}
		result, err := callFunction(n.Name, args)
		if err != nil {
			return nil, err
		}
		trace := fmt.Sprintf("%s(%s) -> %s", n.Name, strings.Join(argTraces, ", "), formatValue(result))
		return &Result{Value: result, Trace: trace}, nil
	}
}

func applyBinaryOp(op string, left, right any) any {
	switch op {
	case "==":
		return equals(left, right)
	case "!=":
		return !equals(left, right)
	case "&&":
		return toBool(left) && toBool(right)
	case "||":
		return toBool(left) || toBool(right)
	case "<":
		return toFloat(left) < toFloat(right)
	case ">":
		return toFloat(left) > toFloat(right)
	case "<=":
		return toFloat(left) <= toFloat(right)
	case ">=":
		return toFloat(left) >= toFloat(right)
	}

	return nil
}

func applyUnaryOp(op string, operand any) any {
	if op == "!" {
		return !toBool(operand)
	}
	return nil
}

func callFunction(name string, args []any) (any, error) {
	switch name {
	case "contains":
		if len(args) != 2 {
			return nil, fmt.Errorf("contains requires 2 arguments")
		}
		return strings.Contains(toString(args[0]), toString(args[1])), nil

	case "startsWith":
		if len(args) != 2 {
			return nil, fmt.Errorf("startsWith requires 2 arguments")
		}
		return strings.HasPrefix(toString(args[0]), toString(args[1])), nil

	case "endsWith":
		if len(args) != 2 {
			return nil, fmt.Errorf("endsWith requires 2 arguments")
		}
		return strings.HasSuffix(toString(args[0]), toString(args[1])), nil

	case "format":
		if len(args) < 1 {
			return nil, fmt.Errorf("format requires at least 1 argument")
		}
		return formatString(toString(args[0], args[1:])), nil

	case "join":
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("join requires 1 or 2 arguments")
		}

		sep := ","
		if len(args) == 2 {
			sep = toString(args[1])
		}

		return joinArray(args[0], sep), nil

	case "always":
		return true, nil
	case "success":
		return true, nil // Assume success for dryruns.
	case "failure":
		return false, nil
	case "cancelled":
		return false, nil
	}

	return nil, fmt.Errorf("unknown function: %s", name)
}

func equals(a, b any) bool {
	return toString(a) == toString(b)
}

func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case nil:
		return false
	case int, int64, float64:
		return val != 0
	}

	return true
}

func toString(v any) string {
	if v == nil {
		return ""
	}

	return fmt.Sprintf("%v", v)
}

func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}

	return 0
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case string:
		return fmt.Sprintf("'%s'", val)
	}

	return fmt.Sprintf("%v", v)
}

func formatString(format string, args []any) string {
	result := format

	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, toString(arg))
	}

	return result
}

func joinArray(v any, sep string) string {
	if arr, ok := v.([]any); ok {
		var strs []string

		for _, item := range arr {
			strs = append(strs, toString(item))
		}

		return strings.Join(strs, sep)
	}

	return toString(v)
}
