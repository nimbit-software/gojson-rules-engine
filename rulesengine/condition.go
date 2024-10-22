package rulesengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Condition represents an individual condition within a rule in the rules engine.
// Conditions can compare facts to values using operators, and they can also nest other conditions.
// Fields:
// - Priority: Optional priority of the condition, must be greater than zero if set.
// - Name: The name of the condition.
// - Operator: The operator to be applied for comparison (e.g., equals, greaterThan).
// - Value: The value to compare the fact to.
// - Fact: The fact that is being evaluated in the condition.
// - FactResult: The result of fact evaluation.
// - Result: The evaluation result of the condition (true/false).
// - Params: Additional parameters that may affect the condition's evaluation.
// - Condition: Raw condition string (for debugging or custom use cases).
// - All, Any: Nested conditions that require all or any of the sub-conditions to be true.
// - Not: A nested condition that negates its result.
type Condition struct {
	Priority   *int
	Name       string
	Operator   string
	Value      ValueNode
	Fact       string
	FactResult Fact
	Result     bool
	Params     map[string]interface{}
	Condition  string
	All        []*Condition
	Any        []*Condition
	Not        *Condition
}

// Validate checks if the Condition is valid based on business rules.
// It verifies that if a value, fact, or operator are set, all three must be set.
// It also ensures that if nested conditions (Any, All, Not) are provided, no value, fact, or operator is set.
// Returns an error if the condition is invalid
func (c *Condition) Validate() error {
	// Validate priority (must be greater than 0 if set)
	if c.Priority != nil && *c.Priority <= 0 {
		return errors.New("priority must be greater than zero")
	}

	valueExists := c.Value.Type != Null || (c.Value.Type != String && c.Value.String != "")
	// Validate that if any of Value, Fact, or Operator are set, all three must be set
	if valueExists || c.Operator != "" || c.Fact != "" {
		if !valueExists || c.Operator == "" || c.Fact == "" {
			return errors.New("if value, operator, or fact are set, all three must be provided")
		}
	}
	// If Any, All, or Not are set, Value, Operator, and Fact must not be set
	if (len(c.Any) > 0 || len(c.All) > 0 || c.Not != nil) && (valueExists || c.Operator != "" || c.Fact != "") {
		return errors.New("value, operator, and fact must not be set if any, all, or not conditions are provided")
	}

	return nil
}

// UnmarshalJSON is a custom JSON unmarshaller for the Condition struct.
// It validates the condition after unmarshalling to ensure it adheres to the rules.
// Params:
// - data: JSON data representing the condition.
// Returns an error if the condition is invalid after unmarshalling.
func (c *Condition) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to hold the incoming data
	type Alias Condition // Alias to avoid infinite recursion inEvaluator UnmarshalJSON
	temp := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	// Unmarshal the JSON data into the temp struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Validate the condition after unmarshaling
	if err := c.Validate(); err != nil {
		return err
	}
	return nil
}

// ToJSON converts the Condition instance to a JSON string representation.
// Useful for serializing the condition for storage or transmission.
func (c *Condition) ToJSON(stringify bool) (interface{}, error) {
	props := map[string]interface{}{}
	if c.Priority != nil {
		props["priority"] = *c.Priority
	}
	if c.Name != "" {
		props["name"] = c.Name
	}
	if oper := c.booleanOperator(); oper != "" {
		if c.All != nil {
			allConditions := make([]interface{}, len(c.All))
			for i, condition := range c.All {
				jsonCondition, err := condition.ToJSON(false)
				if err != nil {
					return nil, err
				}
				allConditions[i] = jsonCondition
			}
			props["all"] = allConditions
		}
		if c.Any != nil {
			anyConditions := make([]interface{}, len(c.Any))
			for i, condition := range c.Any {
				jsonCondition, err := condition.ToJSON(false)
				if err != nil {
					return nil, err
				}
				anyConditions[i] = jsonCondition
			}
			props["any"] = anyConditions
		}
		if c.Not != nil {
			jsonCondition, err := c.Not.ToJSON(false)
			if err != nil {
				return nil, err
			}
			props["not"] = jsonCondition
		}
	} else if c.IsConditionReference() {
		props["condition"] = c.Condition
	} else {
		props["operator"] = c.Operator
		props["value"] = c.Value
		props["fact"] = c.Fact
		props["factResult"] = c.FactResult
		props["result"] = c.Result

		if c.Params != nil {
			props["params"] = c.Params
		}
	}

	if stringify {
		jsonStr, err := json.Marshal(props)
		if err != nil {
			return nil, err
		}
		return string(jsonStr), nil
	}
	return props, nil
}

// Evaluate evaluates the condition against the given almanac and operator map
func (c *Condition) Evaluate(almanac *Almanac, operatorMap map[string]Operator) (*EvaluationResult, error) {
	if reflect.ValueOf(almanac).IsZero() {
		return nil, errors.New("almanac required")
	}
	if reflect.ValueOf(operatorMap).IsZero() {
		return nil, errors.New("operatorMap required")
	}
	if c.IsBooleanOperator() {
		return nil, errors.New("Cannot evaluate() a boolean condition")
	}

	op, ok := operatorMap[c.Operator]
	if !ok {
		return nil, fmt.Errorf("Unknown operator: %s", c.Operator)
	}

	rightHandSideValue := c.Value
	leftHandSideValue, err := almanac.FactValue(c.Fact)
	if err != nil {
		return nil, err
	}

	var result bool
	if leftHandSideValue != nil && leftHandSideValue.Value != nil {
		result = op.Evaluate(leftHandSideValue.Value, &rightHandSideValue)
		// TODO VALUE
		Debug(fmt.Sprintf(`condition::evaluate <%v %s %v?> (%v)`, leftHandSideValue.Value.Raw(), c.Operator, rightHandSideValue, result))
	}

	res := &EvaluationResult{
		Result:             result,
		RightHandSideValue: rightHandSideValue,
		Operator:           c.Operator,
	}
	if leftHandSideValue != nil {
		res.LeftHandSideValue = *leftHandSideValue
	}
	return res, nil
}

// booleanOperator returns the boolean operator for the condition
func booleanOperator(condition *Condition) string {
	if len(condition.Any) > 0 {
		return "any"
	} else if len(condition.All) > 0 {
		return "all"
	} else if condition.Not != nil {
		return "not"
	}
	return ""
}

// booleanOperator returns the condition's boolean operator
func (c *Condition) booleanOperator() string {
	if c == nil {
		return ""
	}
	if c.All != nil {
		return "all"
	}
	if c.Any != nil {
		return "any"
	}
	if c.Not != nil {
		return "not"
	}
	return ""
}

// IsBooleanOperator returns whether the operator is boolean ('all', 'any', 'not')
func (c *Condition) IsBooleanOperator() bool {
	return c.booleanOperator() != ""
}

// isConditionReference returns whether the condition represents a reference to a condition
func (c *Condition) IsConditionReference() bool {
	if c == nil {
		return false
	}
	_, ok := reflect.TypeOf(*c).FieldByName("Condition")
	return ok && c.Condition != ""
}
