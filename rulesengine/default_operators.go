package rulesengine

import (
	"github.com/tidwall/gjson"
	"strings"
)

// NumberValidator checks if the value is a valid number

func equalsEvaluator(a, b gjson.Result) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case gjson.Null:
		return true // Both are null
	case gjson.False, gjson.True:
		return a.Bool() == b.Bool()
	case gjson.Number:
		return a.Num == b.Num
	case gjson.String:
		return a.Str == b.Str
	case gjson.JSON:
		// For arrays and objects, we need to compare their contents
		if a.IsArray() && b.IsArray() {
			aArray := a.Array()
			bArray := b.Array()
			if len(aArray) != len(bArray) {
				return false
			}
			for i := range aArray {
				if !equalsEvaluator(aArray[i], bArray[i]) {
					return false
				}
			}
			return true
		} else if a.IsObject() && b.IsObject() {
			aMap := a.Map()
			bMap := b.Map()
			if len(aMap) != len(bMap) {
				return false
			}
			for key, aVal := range aMap {
				bVal, exists := bMap[key]
				if !exists || !equalsEvaluator(aVal, bVal) {
					return false
				}
			}
			return true
		} else {
			// One is array, the other is object, or some other mismatch
			return false
		}
	default:
		// Unhandled type
		return false
	}
}

// NOT EQUALS OPERATOR
func notEqualsEvaluator(a, b gjson.Result) bool {
	return !equalsEvaluator(a, b)
}

// IN OPERATOR
func inEvaluator(a, b gjson.Result) bool {
	if !b.IsArray() {
		return false
	}
	for _, element := range b.Array() {
		if equalsEvaluator(a, element) {
			return true
		}
	}
	return false
}

// NOT IN OPERATOR
func notInEvaluator(a, b gjson.Result) bool {
	return !inEvaluator(a, b)
}

// CONTAINS OPERATOR
func containsEvaluator(a, b gjson.Result) bool {
	if !a.IsArray() {
		return false
	}
	for _, element := range a.Array() {
		if equalsEvaluator(element, b) {
			return true
		}
	}
	return false
}

// DOES NOT CONTAIN OPERATOR
func doesNotContainEvaluator(a, b gjson.Result) bool {
	return !containsEvaluator(a, b)
}

// LESS THAN OPERATOR
func lessThanEvaluator(a, b gjson.Result) bool {
	return a.Num < b.Num
}

// LESS THAN OPERATOR
func lessThanInclusiveEvaluator(a, b gjson.Result) bool {
	return a.Num <= b.Num
}

// GREATER THAN OPERATOR
func greaterThanEvaluator(a, b gjson.Result) bool {
	return a.Num > b.Num
}

// GREATER THAN OPERATOR
func greaterThanInclusiveEvaluator(a, b gjson.Result) bool {
	return a.Num >= b.Num
}

// STARTS WITH OPERATOR
func startsWithEvaluator(a, b gjson.Result) bool {
	return strings.HasPrefix(a.String(), b.String())
}

// ENDS WITH OPERATOR
func endsWithEvaluator(a, b gjson.Result) bool {
	return strings.HasSuffix(a.String(), b.String())
}

// INCLUDES STRING OPERATOR
func includesStringEvaluator(a, b gjson.Result) bool {
	return strings.Contains(a.String(), b.String())
}

// **************************************************************************************
// FACT VALIDATOR FUNCTIONS
func exists(a gjson.Result) bool {
	return a.Exists()
}

func numberValidator(factValue gjson.Result) bool {
	return factValue.Type == gjson.Number
}

func stringValidator(factValue gjson.Result) bool {
	return factValue.Type == gjson.String
}

// DefaultOperators returns a slice of default operators
func DefaultOperators() []Operator {
	var operators []Operator

	// EQUALS
	equal, _ := NewOperator("equal", equalsEvaluator, nil)
	operators = append(operators, *equal)
	equal, _ = NewOperator("=", equalsEvaluator, nil)
	operators = append(operators, *equal)
	equal, _ = NewOperator("eq", equalsEvaluator, nil)
	operators = append(operators, *equal)

	// NOT EQUALS
	notEqual, _ := NewOperator("notEqual", notEqualsEvaluator, nil)
	operators = append(operators, *notEqual)
	notEqual, _ = NewOperator("ne", notEqualsEvaluator, nil)
	operators = append(operators, *notEqual)
	notEqual, _ = NewOperator("!=", notEqualsEvaluator, nil)
	operators = append(operators, *notEqual)

	// IN OPERATOR
	in, _ := NewOperator("in", inEvaluator, nil)
	operators = append(operators, *in)

	// NOT IN OPERATOR
	notIn, _ := NewOperator("notIn", notInEvaluator, nil)
	operators = append(operators, *notIn)

	// CONTAINS OPERATOR
	contains, _ := NewOperator("contains", containsEvaluator, exists)
	operators = append(operators, *contains)

	// DOES NOT CONTAIN OPERATOR
	notContains, _ := NewOperator("doesNotContain", doesNotContainEvaluator, exists)
	operators = append(operators, *notContains)

	// LESS THAN OPERATOR
	lessThan, _ := NewOperator("lessThan", lessThanEvaluator, numberValidator)
	operators = append(operators, *lessThan)
	lessThan, _ = NewOperator("<", lessThanEvaluator, numberValidator)
	operators = append(operators, *lessThan)
	lessThan, _ = NewOperator("lt", lessThanEvaluator, numberValidator)
	operators = append(operators, *lessThan)

	// LESS THAN INCLUSIVE OPERATOR
	lessThanInclusive, _ := NewOperator("lessThanInclusive", lessThanInclusiveEvaluator, numberValidator)
	operators = append(operators, *lessThanInclusive)
	lessThanInclusive, _ = NewOperator("<=", lessThanInclusiveEvaluator, numberValidator)
	operators = append(operators, *lessThanInclusive)
	lessThanInclusive, _ = NewOperator("lte", lessThanInclusiveEvaluator, numberValidator)
	operators = append(operators, *lessThanInclusive)

	// GREATER THAN OPERATOR
	greaterThan, _ := NewOperator("greaterThan", greaterThanEvaluator, numberValidator)
	operators = append(operators, *greaterThan)
	greaterThan, _ = NewOperator(">", greaterThanEvaluator, numberValidator)
	operators = append(operators, *greaterThan)
	greaterThan, _ = NewOperator("gt", greaterThanEvaluator, numberValidator)
	operators = append(operators, *greaterThan)

	// GREATER THAN INCLUSIVE OPERATOR
	greaterThanInclusive, _ := NewOperator("greaterThanInclusive", greaterThanInclusiveEvaluator, numberValidator)
	operators = append(operators, *greaterThanInclusive)

	greaterThanInclusive, _ = NewOperator(">=", greaterThanInclusiveEvaluator, numberValidator)
	operators = append(operators, *greaterThanInclusive)

	greaterThanInclusive, _ = NewOperator("gte", greaterThanInclusiveEvaluator, numberValidator)
	operators = append(operators, *greaterThanInclusive)

	// STARTS WITH
	startsWith, _ := NewOperator("startsWith", startsWithEvaluator, stringValidator)
	operators = append(operators, *startsWith)

	endsWith, _ := NewOperator("endsWith", endsWithEvaluator, stringValidator)
	operators = append(operators, *endsWith)

	includes, _ := NewOperator("includes", includesStringEvaluator, stringValidator)
	operators = append(operators, *includes)

	return operators
}
