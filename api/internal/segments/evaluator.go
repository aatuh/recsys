package segments

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Evaluator executes JSON rule expressions against a context payload.
type Evaluator struct {
	data map[string]any
	now  time.Time
}

// NewEvaluator constructs an evaluator with the provided context data.
func NewEvaluator(data map[string]any, now time.Time) *Evaluator {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &Evaluator{data: data, now: now}
}

// Match returns true if the rule evaluates to true under the current context.
func (e *Evaluator) Match(rule json.RawMessage) (bool, error) {
	if len(rule) == 0 {
		return false, errors.New("empty rule expression")
	}

	dec := json.NewDecoder(bytes.NewReader(rule))
	dec.UseNumber()
	var expr any
	if err := dec.Decode(&expr); err != nil {
		return false, fmt.Errorf("decode rule: %w", err)
	}

	return e.eval(expr)
}

func (e *Evaluator) eval(node any) (bool, error) {
	switch v := node.(type) {
	case map[string]any:
		if len(v) == 0 {
			return false, nil
		}
		for op, val := range v {
			switch strings.ToLower(op) {
			case "any":
				arr, ok := val.([]any)
				if !ok {
					return false, fmt.Errorf("any expects array, got %T", val)
				}
				for _, child := range arr {
					matched, err := e.eval(child)
					if err != nil {
						return false, err
					}
					if matched {
						return true, nil
					}
				}
				return false, nil
			case "all":
				arr, ok := val.([]any)
				if !ok {
					return false, fmt.Errorf("all expects array, got %T", val)
				}
				for _, child := range arr {
					matched, err := e.eval(child)
					if err != nil {
						return false, err
					}
					if !matched {
						return false, nil
					}
				}
				return true, nil
			case "not":
				matched, err := e.eval(val)
				if err != nil {
					return false, err
				}
				return !matched, nil
			case "eq":
				return e.compare(opEq, val)
			case "neq":
				res, err := e.compare(opEq, val)
				return !res, err
			case "gt":
				return e.compare(opGT, val)
			case "gte":
				return e.compare(opGTE, val)
			case "lt":
				return e.compare(opLT, val)
			case "lte":
				return e.compare(opLTE, val)
			case "in":
				return e.inOperator(val)
			case "contains":
				return e.containsOperator(val)
			case "exists":
				path, ok := val.(string)
				if !ok {
					return false, fmt.Errorf("exists expects string path, got %T", val)
				}
				_, ok = e.resolve(path)
				return ok, nil
			case "gte_days_since":
				return e.gteDaysSince(val)
			default:
				return false, fmt.Errorf("unsupported operator %q", op)
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("rule node must be object, got %T", node)
	}
}

type comparator int

const (
	opEq comparator = iota
	opGT
	opGTE
	opLT
	opLTE
)

func (e *Evaluator) compare(op comparator, val any) (bool, error) {
	params, ok := val.([]any)
	if !ok || len(params) != 2 {
		return false, fmt.Errorf("comparison expects [path, value], got %T", val)
	}
	path, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("comparison path must be string, got %T", params[0])
	}
	actual, ok := e.resolve(path)
	if !ok {
		return false, nil
	}

	switch target := params[1].(type) {
	case json.Number:
		av, ok := toFloat(actual)
		if !ok {
			return false, nil
		}
		tv, err := target.Float64()
		if err != nil {
			return false, err
		}
		return compareFloats(op, av, tv), nil
	case float64, float32, int, int64, int32, uint64, uint32:
		av, ok := toFloat(actual)
		if !ok {
			return false, nil
		}
		tv, _ := toFloat(target)
		return compareFloats(op, av, tv), nil
	case string:
		aStr, ok := toString(actual)
		if !ok {
			return false, nil
		}
		return compareStrings(op, aStr, target), nil
	case bool:
		aBool, ok := toBool(actual)
		if !ok {
			return false, nil
		}
		return compareBools(op, aBool, target), nil
	default:
		return false, fmt.Errorf("unsupported comparison target type %T", target)
	}
}

func compareFloats(op comparator, a, b float64) bool {
	switch op {
	case opEq:
		return a == b
	case opGT:
		return a > b
	case opGTE:
		return a >= b
	case opLT:
		return a < b
	case opLTE:
		return a <= b
	default:
		return false
	}
}

func compareStrings(op comparator, a, b string) bool {
	switch op {
	case opEq:
		return a == b
	case opGT:
		return a > b
	case opGTE:
		return a >= b
	case opLT:
		return a < b
	case opLTE:
		return a <= b
	default:
		return false
	}
}

func compareBools(op comparator, a, b bool) bool {
	switch op {
	case opEq:
		return a == b
	default:
		return false
	}
}

func (e *Evaluator) inOperator(val any) (bool, error) {
	params, ok := val.([]any)
	if !ok || len(params) != 2 {
		return false, fmt.Errorf("in expects [path, values], got %T", val)
	}
	path, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("in path must be string, got %T", params[0])
	}
	actual, ok := e.resolve(path)
	if !ok {
		return false, nil
	}
	set, ok := params[1].([]any)
	if !ok {
		return false, fmt.Errorf("in values must be array, got %T", params[1])
	}

	for _, candidate := range set {
		if equals(actual, candidate) {
			return true, nil
		}
	}
	return false, nil
}

func (e *Evaluator) containsOperator(val any) (bool, error) {
	params, ok := val.([]any)
	if !ok || len(params) != 2 {
		return false, fmt.Errorf("contains expects [path, value], got %T", val)
	}
	path, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("contains path must be string, got %T", params[0])
	}
	actual, ok := e.resolve(path)
	if !ok {
		return false, nil
	}

	switch container := actual.(type) {
	case string:
		needle, ok := toString(params[1])
		if !ok {
			return false, nil
		}
		return strings.Contains(strings.ToLower(container), strings.ToLower(needle)), nil
	case []any:
		for _, item := range container {
			if equals(item, params[1]) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

func (e *Evaluator) gteDaysSince(val any) (bool, error) {
	params, ok := val.([]any)
	if !ok || len(params) != 2 {
		return false, fmt.Errorf("gte_days_since expects [path, days], got %T", val)
	}
	path, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("gte_days_since path must be string, got %T", params[0])
	}
	actual, ok := e.resolve(path)
	if !ok {
		return false, nil
	}
	days, ok := toFloat(params[1])
	if !ok {
		return false, fmt.Errorf("gte_days_since days must be numeric, got %T", params[1])
	}
	ts, ok := toTime(actual)
	if !ok {
		return false, nil
	}
	return e.now.Sub(ts) >= time.Duration(days*24)*time.Hour, nil
}

func (e *Evaluator) resolve(path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = e.data
	for _, part := range parts {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[part]
			if !ok {
				return nil, false
			}
			current = next
		default:
			return nil, false
		}
	}
	return current, true
}

func equals(a, b any) bool {
	switch av := a.(type) {
	case json.Number:
		if bv, ok := toFloat(b); ok {
			fa, _ := toFloat(av)
			return fa == bv
		}
	case float64, float32, int, int32, int64, uint32, uint64:
		fa, _ := toFloat(av)
		fb, ok := toFloat(b)
		return ok && fa == fb
	case string:
		bs, ok := toString(b)
		return ok && strings.EqualFold(av, bs)
	case bool:
		bb, ok := toBool(b)
		return ok && av == bb
	case nil:
		return b == nil
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func toString(v any) (string, bool) {
	switch s := v.(type) {
	case string:
		return s, true
	case json.Number:
		return s.String(), true
	case fmt.Stringer:
		return s.String(), true
	default:
		return "", false
	}
}

func toBool(v any) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		switch strings.ToLower(b) {
		case "true", "1", "yes":
			return true, true
		case "false", "0", "no":
			return false, true
		}
	}
	return false, false
}

func toTime(v any) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case string:
		if ts, err := time.Parse(time.RFC3339, t); err == nil {
			return ts, true
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}
