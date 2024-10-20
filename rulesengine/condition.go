package rulesengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Condition represents a condition in the rule engine
type Condition struct {
	Priority   int
	Name       string
	Operator   string
	Value      interface{}
	Fact       string
	FactResult interface{}
	Result     interface{}
	Params     map[string]interface{}
	Path       string
	Condition  string
	All        []*Condition
	Any        []*Condition
	Not        *Condition
}

func (c *Condition) SetPriority(priority int) {
	c.Priority = priority
}

// NewCondition creates a new Condition instance
func NewCondition(properties map[string]interface{}) (*Condition, error) {
	if properties == nil {
		return nil, errors.New("condition: constructor options required")
	}

	cond := &Condition{}
	booleanOperator := booleanOperator(properties)

	if booleanOperator != "" {
		subConditions, subConditionsIsArray := properties[booleanOperator].([]interface{})
		if booleanOperator != "not" && !subConditionsIsArray {
			return nil, fmt.Errorf(`"%s" must be an array`, booleanOperator)
		}
		if booleanOperator == "not" && subConditionsIsArray {
			return nil, fmt.Errorf(`"%s" cannot be an array`, booleanOperator)
		}

		cond.Operator = booleanOperator

		if priority, err := ParsePriority(properties); err == nil {
			cond.Priority = priority
		} else if err.Code == "INVALID_PRIORITY_TYPE" || err.Code == "INVALID_PRIORITY_VALUE" {
			return nil, err
		} else {
			cond.Priority = 1
		}

		if subConditionsIsArray {
			for _, sc := range subConditions {
				subCondMap, ok := sc.(map[string]interface{})
				if !ok {
					return nil, errors.New("invalid sub-condition format")
				}
				subCond, err := NewCondition(subCondMap)
				if err != nil {
					return nil, err
				}
				if booleanOperator == "all" {
					cond.All = append(cond.All, subCond)
				} else {
					cond.Any = append(cond.Any, subCond)
				}
			}
		} else {
			subCondMap, ok := properties[booleanOperator].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid sub-condition format")
			}
			subCond, err := NewCondition(subCondMap)
			if err != nil {
				return nil, err
			}
			cond.Not = subCond
		}
	} else {
		if _, ok := properties["fact"]; !ok {
			return nil, errors.New(`condition: constructor "fact" property required`)
		}
		if _, ok := properties["operator"]; !ok {
			return nil, errors.New(`condition: constructor "operator" property required`)
		}
		if _, ok := properties["value"]; !ok {
			return nil, errors.New(`condition: constructor "value" property required`)
		}

		cond.Fact = properties["fact"].(string)
		if path, ok := properties["path"]; ok {
			cond.Path = path.(string)
		} else {
			cond.Path = ""
		}
		cond.Operator = properties["operator"].(string)
		cond.Value = properties["value"]
		if priority, err := ParsePriority(properties); err == nil {
			cond.Priority = priority
		} else if err.Code == "INVALID_PRIORITY_TYPE" || err.Code == "INVALID_PRIORITY_VALUE" {
			return nil, err
		}
	}

	return cond, nil
}

// ToJSON converts the condition to a JSON-friendly structure
func (c *Condition) ToJSON(stringify bool) (interface{}, error) {
	props := map[string]interface{}{}
	if c.Priority != 0 {
		props["priority"] = c.Priority
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
		if c.FactResult != nil {
			props["factResult"] = c.FactResult
		}
		if c.Result != nil {
			props["result"] = c.Result
		}
		if c.Params != nil {
			props["params"] = c.Params
		}
		if c.Path != "" {
			props["path"] = c.Path
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

	rightHandSideValue, err := almanac.GetValue(c.Value)
	if err != nil {
		return nil, err
	}
	leftHandSideValue, err := almanac.FactValue(c.Fact, c.Params, c.Path)
	if err != nil {
		return nil, err
	}

	result := op.Evaluate(leftHandSideValue, rightHandSideValue)
	Debug(fmt.Sprintf(`condition::evaluate <%v %s %v?> (%v)`, leftHandSideValue, c.Operator, rightHandSideValue, result))

	return &EvaluationResult{
		Result:             result,
		LeftHandSideValue:  leftHandSideValue,
		RightHandSideValue: rightHandSideValue,
		Operator:           c.Operator,
	}, nil
}

// booleanOperator returns the boolean operator for the condition
func booleanOperator(condition map[string]interface{}) string {
	if _, ok := condition["any"]; ok {
		return "any"
	} else if _, ok := condition["all"]; ok {
		return "all"
	} else if _, ok := condition["not"]; ok {
		return "not"
	}
	return ""
}

// booleanOperator returns the condition's boolean operator
func (c *Condition) booleanOperator() string {
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
