package rulesengine

import (
	"encoding/json"
	"testing"
)

func TestCondition(t *testing.T) {

	// Test a valid RuleConfig with a valid Condition
	t.Run("TestValidRuleConfig", func(t *testing.T) {
		priority := 1
		ruleConfig := RuleConfig{
			Name:     "Test Rule",
			Priority: nil, // optional priority
			Conditions: Condition{
				Priority: &priority,
				Operator: "equal",
				Fact:     "factName",
				Value:    "someValue",
			},
			Event: EventConfig{Type: "TestEvent"},
		}

		if err := ruleConfig.Conditions.Validate(); err != nil {
			t.Errorf("Expected RuleConfig to be valid, but got error: %v", err)
		}
	})

	// Test that RuleConfig returns an error when Condition's priority is invalid
	t.Run("TestRuleConfigInvalidPriority", func(t *testing.T) {
		priority := 0
		ruleConfig := RuleConfig{
			Name: "Test Rule",
			Conditions: Condition{
				Priority: &priority,
				Operator: "equal",
				Fact:     "factName",
				Value:    "someValue",
			},
			Event: EventConfig{Type: "TestEvent"},
		}

		err := ruleConfig.Conditions.Validate()
		if err == nil || err.Error() != "priority must be greater than zero" {
			t.Errorf("Expected priority validation error, but got: %v", err)
		}
	})

	// Test that RuleConfig returns an error when Value, Fact, or Operator are missing
	t.Run(" TestRuleConfigMissingValueFactOperator", func(t *testing.T) {
		priority := 1
		testCases := []struct {
			name       string
			conditions Condition
			errMsg     string
		}{
			{
				name: "Missing Fact",
				conditions: Condition{
					Priority: &priority,
					Operator: "equal",
					Value:    "someValue",
					Fact:     "", // missing fact
				},
				errMsg: "if value, operator, or fact are set, all three must be provided",
			},
			{
				name: "Missing Operator",
				conditions: Condition{
					Priority: &priority,
					Operator: "",
					Value:    "someValue",
					Fact:     "factName", // missing operator
				},
				errMsg: "if value, operator, or fact are set, all three must be provided",
			},
			{
				name: "Missing Value",
				conditions: Condition{
					Priority: &priority,
					Operator: "equal",
					Value:    nil, // missing value
					Fact:     "factName",
				},
				errMsg: "if value, operator, or fact are set, all three must be provided",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ruleConfig := RuleConfig{
					Name:       "Test Rule",
					Conditions: tc.conditions,
					Event:      EventConfig{Type: "TestEvent"},
				}

				err := ruleConfig.Conditions.Validate()
				if err == nil || err.Error() != tc.errMsg {
					t.Errorf("Expected error: %v, but got: %v", tc.errMsg, err)
				}
			})
		}
	})

	// Test mutual exclusion of Any, All, and Not with Value, Fact, and Operator
	t.Run("TestRuleConfigMutualExclusion", func(t *testing.T) {
		priority := 1
		ruleConfig := RuleConfig{
			Name: "Test Rule",
			Conditions: Condition{
				Priority: &priority,
				Operator: "equal",
				Fact:     "factName",
				Value:    "someValue",
				All:      []*Condition{{Priority: &priority}}, // All is set, but Value, Fact, Operator are also set
			},
			Event: EventConfig{Type: "TestEvent"},
		}

		err := ruleConfig.Conditions.Validate()
		if err == nil || err.Error() != "value, operator, and fact must not be set if any, all, or not conditions are provided" {
			t.Errorf("Expected mutual exclusion validation error, but got: %v", err)
		}
	})

	// Test that Path can only be set if Value is provided
	t.Run("TestRuleConfigPathRequiresValue", func(t *testing.T) {
		priority := 1
		ruleConfig := RuleConfig{
			Name: "Test Rule",
			Conditions: Condition{
				Priority: &priority,
				Path:     "somePath", // Path is set, but Value is nil
			},
			Event: EventConfig{Type: "TestEvent"},
		}

		err := ruleConfig.Conditions.Validate()
		if err == nil || err.Error() != "path can only be set if value is provided" {
			t.Errorf("Expected path validation error, but got: %v", err)
		}
	})

	// Test unmarshalling valid RuleConfig JSON
	t.Run("TestUnmarshalValidRuleConfig", func(t *testing.T) {
		jsonData := []byte(`{
        "name": "Test Rule",
        "conditions": {
            "priority": 1,
            "operator": "equal",
            "fact": "factName",
            "value": "someValue"
        },
        "event": {
            "type": "TestEvent"
        }
    }`)

		var ruleConfig RuleConfig
		err := json.Unmarshal(jsonData, &ruleConfig)
		if err != nil {
			t.Errorf("Expected successful unmarshal, but got error: %v", err)
		}

		if *ruleConfig.Conditions.Priority != 1 {
			t.Errorf("Expected priority to be 1, got %d", ruleConfig.Conditions.Priority)
		}
		if ruleConfig.Conditions.Operator != "equal" {
			t.Errorf("Expected operator to be 'equal', got %s", ruleConfig.Conditions.Operator)
		}
	})

	// Test unmarshalling invalid RuleConfig JSON (missing fact)
	t.Run("TestUnmarshalInvalidRuleConfig", func(t *testing.T) {
		jsonData := []byte(`{
        "name": "Test Rule",
        "conditions": {
            "priority": 1,
            "operator": "equal",
            "value": "someValue"
        },
        "event": {
            "type": "TestEvent"
        }
    }`)

		var ruleConfig RuleConfig
		err := json.Unmarshal(jsonData, &ruleConfig)
		if err == nil || err.Error() != "if value, operator, or fact are set, all three must be provided" {
			t.Errorf("Expected unmarshalling error for missing fact, but got: %v", err)
		}
	})
}
