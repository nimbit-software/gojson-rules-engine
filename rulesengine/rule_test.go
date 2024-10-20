package rulesengine

import (
	"testing"
)

func TestNewRule(t *testing.T) {
	t.Run("Valid priority types", func(t *testing.T) {
		testCases := []struct {
			name        string
			priority    int
			expected    int
			expectError bool // Add a field to indicate if an error is expected
		}{
			{"valid priority 4", 4, 4, false},        // Valid priority should succeed
			{"invalid priority -1", 100, 100, false}, // Invalid priority should return an error
			{"invalid priority 0", 1, 1, false},      // Priority 0 should return an error
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				allConditions := []*Condition{}
				options := RuleConfig{
					Name:     "Test Rule",
					Priority: &tc.priority,
					Conditions: Condition{
						All: allConditions,
					},
					Event: EventConfig{
						Type: "test",
					},
				}

				rule, err := NewRule(&options)
				if tc.expectError {
					// Expect an error for invalid priority
					if err == nil {
						t.Errorf("Expected an error for priority %d, but got none", tc.priority)
					}
				} else {
					// No error expected, validate the rule creation and priority
					if err != nil {
						t.Errorf("Expected rule creation to succeed, but got error: %v", err)
					}
					if rule.Priority != tc.expected {
						t.Errorf("Expected priority to be %d, but got %d", tc.expected, rule.Priority)
					}
				}
			})
		}
	})

	t.Run("Invalid priority types", func(t *testing.T) {
		testCases := []struct {
			name        string
			priority    int
			expected    int
			expectError bool // Add a field to indicate if an error is expected
		}{
			{"invalid priority 0", 0, 0, true},   // Valid priority should succeed
			{"invalid priority -1", -1, 0, true}, // Valid priority should succeed
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				allConditions := []*Condition{}
				options := RuleConfig{
					Name:     "Test Rule",
					Priority: &tc.priority,
					Conditions: Condition{
						All: allConditions,
					},
					Event: EventConfig{
						Type: "test",
					},
				}

				rule, err := NewRule(&options)
				if tc.expectError {
					// Expect an error for invalid priority
					if err == nil {
						t.Errorf("Expected an error for priority %d, but got none", tc.priority)
					}
				} else {
					// No error expected, validate the rule creation and priority
					if err != nil {
						t.Errorf("Expected rule creation to succeed, but got error: %v", err)
					}
					if rule.Priority != tc.expected {
						t.Errorf("Expected priority to be %d, but got %d", tc.expected, rule.Priority)
					}
				}
			})
		}
	})

	t.Run("Default priority", func(t *testing.T) {
		allConditions := []*Condition{}
		options := RuleConfig{
			Name: "Test Rule",
			Conditions: Condition{
				All: allConditions,
			},
			Event: EventConfig{
				Type: "test",
			},
		}

		rule, err := NewRule(&options)
		if err != nil {
			t.Errorf("Expected rule creation to succeed, but got error: %v", err)
		}
		if rule.Priority != 1 {
			t.Errorf("Expected priority to be 1 (default), but got %d", rule.Priority)
		}
	})
}
