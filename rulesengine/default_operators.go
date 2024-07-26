package rulesengine

import "strings"

// NumberValidator checks if the value is a valid number
func NumberValidator(factValue interface{}) bool {
	_, ok := factValue.(float64)
	return ok
}

// DefaultOperators returns a slice of default operators
func DefaultOperators() []Operator {
	var operators []Operator

	// EQUALS
	equal, _ := NewOperator("equal", func(a, b interface{}) bool {
		return a == b
	}, nil)
	operators = append(operators, *equal)
	equal, _ = NewOperator("=", func(a, b interface{}) bool {
		return a == b
	}, nil)
	operators = append(operators, *equal)
	equal, _ = NewOperator("eq", func(a, b interface{}) bool {
		return a == b
	}, nil)
	operators = append(operators, *equal)

	// NOT EQUALS
	notEqual, _ := NewOperator("notEqual", func(a, b interface{}) bool {
		return a != b
	}, nil)
	operators = append(operators, *notEqual)
	notEqual, _ = NewOperator("ne", func(a, b interface{}) bool {
		return a != b
	}, nil)
	operators = append(operators, *notEqual)
	notEqual, _ = NewOperator("!=", func(a, b interface{}) bool {
		return a != b
	}, nil)
	operators = append(operators, *notEqual)

	// IN OPERATOR
	in, _ := NewOperator("in", func(a, b interface{}) bool {
		bArray, ok := b.([]interface{})
		if !ok {
			return false
		}
		for _, v := range bArray {
			if v == a {
				return true
			}
		}
		return false
	}, nil)
	operators = append(operators, *in)

	// NOT IN OPERATOR
	notIn, _ := NewOperator("notIn", func(a, b interface{}) bool {
		bArray, ok := b.([]interface{})
		if !ok {
			return true
		}
		for _, v := range bArray {
			if v == a {
				return false
			}
		}
		return true
	}, nil)
	operators = append(operators, *notIn)

	// CONTAINS OPERATOR
	contains, _ := NewOperator("contains", func(a, b interface{}) bool {
		aArray, ok := a.([]interface{})
		if !ok {
			return false
		}
		for _, v := range aArray {
			if v == b {
				return true
			}
		}
		return false
	}, func(factValue interface{}) bool {
		_, ok := factValue.([]interface{})
		return ok
	})
	operators = append(operators, *contains)

	// DOES NOT CONTAIN OPERATOR
	notContains, _ := NewOperator("doesNotContain", func(a, b interface{}) bool {
		aArray, ok := a.([]interface{})
		if !ok {
			return true
		}
		for _, v := range aArray {
			if v == b {
				return false
			}
		}
		return true
	}, func(factValue interface{}) bool {
		_, ok := factValue.([]interface{})
		return ok
	})
	operators = append(operators, *notContains)

	// LESS THAN OPERATOR
	lessThan, _ := NewOperator("lessThan", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat < bFloat
	}, NumberValidator)
	operators = append(operators, *lessThan)

	lessThan, _ = NewOperator("<", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat < bFloat
	}, NumberValidator)
	operators = append(operators, *lessThan)

	lessThan, _ = NewOperator("lt", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat < bFloat
	}, NumberValidator)
	operators = append(operators, *lessThan)

	// LESS THAN INCLUSIVE OPERATOR
	lessThanInclusive, _ := NewOperator("lessThanInclusive", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat <= bFloat
	}, NumberValidator)
	operators = append(operators, *lessThanInclusive)

	lessThanInclusive, _ = NewOperator("<=", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat <= bFloat
	}, NumberValidator)
	operators = append(operators, *lessThanInclusive)

	lessThanInclusive, _ = NewOperator("lte", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat <= bFloat
	}, NumberValidator)
	operators = append(operators, *lessThanInclusive)

	// GREATER THAN OPERATOR
	greaterThan, _ := NewOperator("greaterThan", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat > bFloat
	}, NumberValidator)
	operators = append(operators, *greaterThan)

	greaterThan, _ = NewOperator(">", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat > bFloat
	}, NumberValidator)
	operators = append(operators, *greaterThan)

	greaterThan, _ = NewOperator("gt", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat > bFloat
	}, NumberValidator)
	operators = append(operators, *greaterThan)
	// GREATER THAN INCLUSIVE OPERATOR
	greaterThanInclusive, _ := NewOperator("greaterThanInclusive", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat >= bFloat
	}, NumberValidator)
	operators = append(operators, *greaterThanInclusive)

	greaterThanInclusive, _ = NewOperator(">=", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat >= bFloat
	}, NumberValidator)
	operators = append(operators, *greaterThanInclusive)

	greaterThanInclusive, _ = NewOperator("gte", func(a, b interface{}) bool {
		aFloat, okA := a.(float64)
		bFloat, okB := b.(float64)
		return okA && okB && aFloat >= bFloat
	}, NumberValidator)
	operators = append(operators, *greaterThanInclusive)

	// STARTS WITH
	startsWith, _ := NewOperator("startsWith", func(a, b interface{}) bool {
		aString, okA := a.(string)
		bString, okB := b.(string)
		return okA && okB && strings.HasPrefix(aString, bString)
	}, nil)
	operators = append(operators, *startsWith)

	endsWith, _ := NewOperator("endsWith", func(a, b interface{}) bool {
		aString, okA := a.(string)
		bString, okB := b.(string)
		return okA && okB && strings.HasSuffix(aString, bString)
	}, nil)
	operators = append(operators, *endsWith)

	includes, _ := NewOperator("includes", func(a, b interface{}) bool {
		aString, okA := a.(string)
		bString, okB := b.(string)
		return okA && okB && strings.Contains(aString, bString)
	}, nil)
	operators = append(operators, *includes)

	return operators
}
