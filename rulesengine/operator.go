package rulesengine

import (
	"errors"
)

// Operator represents an operator in the rule engine
type Operator struct {
	Name               string
	Callback           func(factValue, jsonValue interface{}) bool
	FactValueValidator func(factValue interface{}) bool
}

// NewOperator creates a new Operator instance
func NewOperator(name string, cb func(factValue, jsonValue interface{}) bool, factValueValidator func(factValue interface{}) bool) (*Operator, error) {
	if name == "" {
		return nil, errors.New("Missing operator name")
	}
	if cb == nil {
		return nil, errors.New("Missing operator callback")
	}
	if factValueValidator == nil {
		factValueValidator = func(factValue interface{}) bool { return true }
	}
	return &Operator{
		Name:               name,
		Callback:           cb,
		FactValueValidator: factValueValidator,
	}, nil
}

// Evaluate takes the fact result and compares it to the condition 'value' using the callback
func (o *Operator) Evaluate(factValue, jsonValue interface{}) bool {
	return o.FactValueValidator(factValue) && o.Callback(factValue, jsonValue)
}
