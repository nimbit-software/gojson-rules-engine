package rulesengine

import (
	"strings"
)

// EvalEqual checks if two ValueNode instances are equal.
// It compares their types first, and if they match, it evaluates their values.
// Supported types: String, Number, Bool, Array.
// Returns true if both nodes have the same type and value, false otherwise.
func EvalEqual(a, b *ValueNode) bool {
	if !a.SameType(b) {
		return false
	}
	switch a.Type {
	case String:
		return a.String == b.String
	case Number:
		return a.Number == b.Number
	case Bool:
		return a.Bool == b.Bool
	case Array:
		if len(a.Array) != len(b.Array) {
			return false
		}
		for i := range a.Array {
			if !EvalEqual(&a.Array[i], &b.Array[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// EvalNotEquals checks if two ValueNode instances are not equal.
// It returns the negation of the EvalEqual function.
// Returns true if the nodes are not equal, false otherwise.
func EvalNotEquals(a, b *ValueNode) bool {
	return !EvalEqual(a, b)
}

// EvalIn checks if a ValueNode instance is present in an array of ValueNode instances.
// It assumes that 'b' is an array and iterates through it to find a match with 'a'.
// Returns true if 'a' is found in 'b', false otherwise.
func EvalIn(a, b *ValueNode) bool {
	// 'b' should be an array
	if !b.IsArray() {
		return false
	}

	for _, item := range b.Array {
		if a.Type != item.Type {
			continue // Skip if types don't match
		}
		equal := EvalEqual(a, &item)
		if equal {
			return true
		}
	}

	return false
}

// EvalNotIn checks if a ValueNode instance is not present in an array of ValueNode instances.
// It returns the negation of EvalIn.
// Returns true if 'a' is not found in 'b', false otherwise.
func EvalNotIn(a, b *ValueNode) bool {
	return !EvalIn(a, b)
}

// EvalLessThan checks if the first ValueNode is less than the second.
// Both 'a' and 'b' must be numbers for the comparison to be valid.
// Returns true if 'a' is less than 'b', false otherwise.
func EvalLessThan(a, b *ValueNode) bool {
	if !a.IsNumber() || !b.IsNumber() {
		return false
	}
	return a.Number < b.Number
}

// EvalLessThanOrEqual checks if the first ValueNode is less than or equal to the second.
// Both 'a' and 'b' must be numbers for the comparison to be valid.
// Returns true if 'a' is less than or equal to 'b', false otherwise.
func EvalLessThanOrEqual(a, b *ValueNode) bool {
	if !a.IsNumber() || !b.IsNumber() {
		return false
	}
	return a.Number <= b.Number
}

// EvalGreaterThan checks if the first ValueNode is greater than the second.
// Both 'a' and 'b' must be numbers for the comparison to be valid.
// Returns true if 'a' is greater than 'b', false otherwise.
func EvalGreaterThan(a, b *ValueNode) bool {
	if !a.IsNumber() || !b.IsNumber() {
		return false
	}
	return a.Number > b.Number
}

// EvalGreaterOrEqual checks if the first ValueNode is greater than or equal to the second.
// Both 'a' and 'b' must be numbers for the comparison to be valid.
// Returns true if 'a' is greater than or equal to 'b', false otherwise.
func EvalGreaterOrEqual(a, b *ValueNode) bool {
	if !a.IsNumber() || !b.IsNumber() {
		return false
	}
	return a.Number >= b.Number
}

// EvalStartsWith checks if the string in the first ValueNode starts with the string in the second ValueNode.
// Both 'a' and 'b' must be strings for the comparison to be valid.
// Returns true if 'a' starts with 'b', false otherwise.
func EvalStartsWith(a, b *ValueNode) bool {
	if !a.IsString() || !b.IsString() {
		return false
	}
	return strings.HasPrefix(a.String, b.String)
}

// EvalEndsWith checks if the string in the first ValueNode ends with the string in the second ValueNode.
// Both 'a' and 'b' must be strings for the comparison to be valid.
// Returns true if 'a' ends with 'b', false otherwise.
func EvalEndsWith(a, b *ValueNode) bool {
	if !a.IsString() || !b.IsString() {
		return false
	}
	return strings.HasSuffix(a.String, b.String)
}

// EvalIncludes checks if the string in the first ValueNode contains with the string in the second ValueNode.
// Both 'a' and 'b' must be strings for the comparison to be valid.
// Returns true if 'a' ends with 'b', false otherwise.
func EvalIncludes(a, b *ValueNode) bool {
	if !a.IsString() || !b.IsString() {
		return false
	}
	return strings.Contains(a.String, b.String)
}

// **************************************************************************************
// FACT VALIDATOR FUNCTIONS
func exists(a *ValueNode) bool {
	return a.Type != Null
}

func isArray(a *ValueNode) bool {
	return a.Type != Null && a.IsArray()
}

func numberValidator(a *ValueNode) bool {
	return a.Type == Number
}

func stringValidator(a *ValueNode) bool {
	return a.Type == String
}

// DefaultOperators returns a slice of default operators
func DefaultOperators() []Operator {
	var operators []Operator

	// EQUALS
	equal, _ := NewOperator("equal", EvalEqual, nil)
	operators = append(operators, *equal)
	equal, _ = NewOperator("=", EvalEqual, nil)
	operators = append(operators, *equal)
	equal, _ = NewOperator("eq", EvalEqual, nil)
	operators = append(operators, *equal)

	// NOT EQUALS
	notEqual, _ := NewOperator("notEqual", EvalNotEquals, nil)
	operators = append(operators, *notEqual)
	notEqual, _ = NewOperator("ne", EvalNotEquals, nil)
	operators = append(operators, *notEqual)
	notEqual, _ = NewOperator("!=", EvalNotEquals, nil)
	operators = append(operators, *notEqual)

	// IN OPERATOR
	in, _ := NewOperator("in", EvalIn, isArray)
	operators = append(operators, *in)

	// NOT IN OPERATOR
	notIn, _ := NewOperator("notIn", EvalNotIn, isArray)
	operators = append(operators, *notIn)

	// CONTAINS OPERATOR
	contains, _ := NewOperator("contains", EvalIn, isArray)
	operators = append(operators, *contains)

	// DOES NOT CONTAIN OPERATOR
	notContains, _ := NewOperator("doesNotContain", EvalNotIn, isArray)
	operators = append(operators, *notContains)

	// LESS THAN OPERATOR
	lessThan, _ := NewOperator("lessThan", EvalLessThan, numberValidator)
	operators = append(operators, *lessThan)
	lessThan, _ = NewOperator("<", EvalLessThan, numberValidator)
	operators = append(operators, *lessThan)
	lessThan, _ = NewOperator("lt", EvalLessThan, numberValidator)
	operators = append(operators, *lessThan)

	// LESS THAN INCLUSIVE OPERATOR
	lessThanInclusive, _ := NewOperator("lessThanInclusive", EvalLessThanOrEqual, numberValidator)
	operators = append(operators, *lessThanInclusive)
	lessThanInclusive, _ = NewOperator("<=", EvalLessThanOrEqual, numberValidator)
	operators = append(operators, *lessThanInclusive)
	lessThanInclusive, _ = NewOperator("lte", EvalLessThanOrEqual, numberValidator)
	operators = append(operators, *lessThanInclusive)

	// GREATER THAN OPERATOR
	greaterThan, _ := NewOperator("greaterThan", EvalGreaterThan, numberValidator)
	operators = append(operators, *greaterThan)
	greaterThan, _ = NewOperator(">", EvalGreaterThan, numberValidator)
	operators = append(operators, *greaterThan)
	greaterThan, _ = NewOperator("gt", EvalGreaterThan, numberValidator)
	operators = append(operators, *greaterThan)

	// GREATER THAN INCLUSIVE OPERATOR
	greaterThanInclusive, _ := NewOperator("greaterThanInclusive", EvalGreaterOrEqual, numberValidator)
	operators = append(operators, *greaterThanInclusive)

	greaterThanInclusive, _ = NewOperator(">=", EvalGreaterOrEqual, numberValidator)
	operators = append(operators, *greaterThanInclusive)

	greaterThanInclusive, _ = NewOperator("gte", EvalGreaterOrEqual, numberValidator)
	operators = append(operators, *greaterThanInclusive)

	// STARTS WITH
	startsWith, _ := NewOperator("startsWith", EvalStartsWith, stringValidator)
	operators = append(operators, *startsWith)

	endsWith, _ := NewOperator("endsWith", EvalEndsWith, stringValidator)
	operators = append(operators, *endsWith)

	includes, _ := NewOperator("includes", EvalIncludes, stringValidator)
	operators = append(operators, *includes)

	return operators
}
