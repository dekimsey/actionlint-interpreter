package expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rhysd/actionlint"
)

type funcDef struct {
	// argsCount is the number of required arguments. Positive values have to be matched exactly,
	// negative values indicate the abs(minimum) number of arguments required
	argsCount int

	call func(args ...*EvaluationResult) *EvaluationResult
}

func contains(args ...*EvaluationResult) *EvaluationResult {
	// Returns true if search contains item. If search is an array, this function returns true if the item is
	// an element in the array. If search is a string, this function returns true if the item is a substring of
	// search. This function is not case sensitive. Casts values to a string.
	// https://docs.github.com/en/actions/learn-github-actions/expressions#contains
	if len(args) != 2 {
		panic("contains() requires exactly 2 arguments")
	}
	left := args[0]
	right := args[1]

	// src := `contains(github.event.client_payload.payload.repo, 'groot')`
	if left.Primitive() {
		if !right.Primitive() {
			return &EvaluationResult{false, &actionlint.BoolType{}}
		}

		ls := left.CoerceString()
		rs := right.CoerceString()

		// Expression string comparisons are string insensitive
		return &EvaluationResult{
			Value: strings.Contains(ls, rs),
			Type:  &actionlint.BoolType{},
		}
	}
	switch left.Type.(type) {
	case *actionlint.ObjectType:
		return &EvaluationResult{false, &actionlint.BoolType{}}
	case *actionlint.ArrayType:
		if !right.Primitive() { // Can only check for basic types in an array
			return &EvaluationResult{false, &actionlint.BoolType{}}
		}
		s := left.CoerceSlice()
		if s == nil {
			return &EvaluationResult{false, &actionlint.BoolType{}}
		}
		for _, v := range s {
			if right.Equals(v) {
				return &EvaluationResult{true, &actionlint.BoolType{}}
			}
		}
		return &EvaluationResult{false, &actionlint.BoolType{}}
	default:
		return &EvaluationResult{false, &actionlint.BoolType{}}
	}
}

var functions map[string]funcDef = map[string]funcDef{
	"contains": {
		argsCount: 2,
		call:      contains,
	},

	"startswith": {
		argsCount: 2,
		call:      startswith,
	},

	"endswith": {
		argsCount: 2,
		call:      endswith,
	},

	"join": {
		argsCount: -1,
		call:      join,
	},

	"fromjson": {
		argsCount: 1,
		call:      fromjson,
	},
}

func fromjson(args ...*EvaluationResult) *EvaluationResult {
	input := args[0]
	inputStr := input.CoerceString()

	var v any
	if err := json.Unmarshal([]byte(inputStr), &v); err != nil {
		panic(fmt.Errorf("unable to unmarshal `%s` fromjson: %w", inputStr, err))
	}

	switch v.(type) {
	case []any:
		return &EvaluationResult{v, &actionlint.ArrayType{}}
	case map[string]any:
		return &EvaluationResult{v, &actionlint.ObjectType{}}
	case string:
		return &EvaluationResult{v, &actionlint.StringType{}}
	case float64:
		return &EvaluationResult{v, &actionlint.NumberType{}}
	case bool:
		return &EvaluationResult{v, &actionlint.BoolType{}}
	default:
		panic(fmt.Errorf("unknown type %T in fromjson", v))
	}
}

func join(args ...*EvaluationResult) *EvaluationResult {
	separator := ","

	// String
	if args[0].Primitive() {
		return args[0]
	}

	if len(args) > 1 {
		separator = args[1].CoerceString()
	}

	ar := args[0].Value.([]interface{})

	v := make([]string, len(ar))
	for i, a := range ar {
		ar := &EvaluationResult{a, getExprType(a)}
		v[i] = ar.CoerceString()
	}

	return &EvaluationResult{strings.Join(v, separator), &actionlint.StringType{}}
}

func endswith(args ...*EvaluationResult) *EvaluationResult {
	// TODO: Check types of parameters
	left := args[0]
	if !left.Primitive() {
		return &EvaluationResult{false, &actionlint.BoolType{}}
	}

	right := args[1]
	if !left.Primitive() {
		return &EvaluationResult{false, &actionlint.BoolType{}}
	}

	ls := left.CoerceString()
	rs := right.CoerceString()

	// Expression string comparisons are string insensitive
	return &EvaluationResult{strings.HasSuffix(strings.ToLower(ls), strings.ToLower(rs)), &actionlint.BoolType{}}
}

func startswith(args ...*EvaluationResult) *EvaluationResult {
	// TODO: Check types of parameters
	left := args[0]
	if !left.Primitive() {
		return &EvaluationResult{false, &actionlint.BoolType{}}
	}

	right := args[1]
	if !right.Primitive() {
		return &EvaluationResult{false, &actionlint.BoolType{}}
	}

	ls := left.CoerceString()
	rs := right.CoerceString()

	// Expression string comparisons are string insensitive
	return &EvaluationResult{strings.HasPrefix(strings.ToLower(ls), strings.ToLower(rs)), &actionlint.BoolType{}}
}
