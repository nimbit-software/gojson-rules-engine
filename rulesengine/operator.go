package rulesengine

import (
	"errors"
)

// Operator defines a function that compares two ValueNodes and returns a boolean result.
// Operators are used in conditions to perform comparisons like equals, greater than, etc.
type Operator struct {
	Name               string
	Callback           func(a, b *ValueNode) bool
	FactValueValidator func(factValue *ValueNode) bool
}

// NewOperator adds a new operator to the engine.
// Params:
// - name: The name of the operator.
// - op: The operator function to be added.
func NewOperator(name string, cb func(a, b *ValueNode) bool, factValueValidator func(factValue *ValueNode) bool) (*Operator, error) {
	if name == "" {
		return nil, errors.New("Missing operator name")
	}
	if cb == nil {
		return nil, errors.New("Missing operator callback")
	}
	if factValueValidator == nil {
		factValueValidator = func(factValue *ValueNode) bool { return true }
	}
	return &Operator{
		Name:               name,
		Callback:           cb,
		FactValueValidator: factValueValidator,
	}, nil
}

// Evaluate takes the fact result and compares it to the condition 'value' using the callback function.
// Params:
// - a: The fact value.
// - b: The condition value.
// Returns true if the condition is met, false otherwise.
func (o *Operator) Evaluate(a, b *ValueNode) bool {
	return o.FactValueValidator(a) && o.Callback(a, b)
}
