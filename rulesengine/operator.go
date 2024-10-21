package rulesengine

import (
	"errors"
	"github.com/tidwall/gjson"
)

// Operator represents an operator inEvaluator the rule engine
type Operator struct {
	Name               string
	Callback           func(factValue, jsonValue gjson.Result) bool
	FactValueValidator func(factValue gjson.Result) bool
}

// NewOperator creates a new Operator instance
func NewOperator(name string, cb func(factValue, jsonValue gjson.Result) bool, factValueValidator func(factValue gjson.Result) bool) (*Operator, error) {
	if name == "" {
		return nil, errors.New("Missing operator name")
	}
	if cb == nil {
		return nil, errors.New("Missing operator callback")
	}
	if factValueValidator == nil {
		factValueValidator = func(factValue gjson.Result) bool { return true }
	}
	return &Operator{
		Name:               name,
		Callback:           cb,
		FactValueValidator: factValueValidator,
	}, nil
}

// Evaluate takes the fact result and compares it to the condition 'value' using the callback
func (o *Operator) Evaluate(factValue, jsonValue gjson.Result) bool {
	return o.FactValueValidator(factValue) && o.Callback(factValue, jsonValue)
}
