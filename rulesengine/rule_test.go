package rulesengine

import (
	"testing"
)

func TestNewRule(t *testing.T) {
	t.Run("Valid priority types", func(t *testing.T) {
		testCases := []struct {
			name     string
			priority interface{}
			expected int
		}{
			{"float64", float64(2.0), 2},
			{"float", float32(2.0), 2},
			{"int64", int64(3), 3},
			{"int", 4, 4},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := map[string]interface{}{
					"name":     "Test Rule",
					"priority": tc.priority,
					"conditions": map[string]interface{}{
						"all": []interface{}{},
					},
				}

				rule, err := NewRule(options)
				if err != nil {
					t.Errorf("Expected rule creation to succeed, but got error: %v", err)
				}
				if rule.Priority != tc.expected {
					t.Errorf("Expected priority to be %d, but got %d", tc.expected, rule.Priority)
				}
			})
		}
	})

	t.Run("Invalid priority types", func(t *testing.T) {
		testCases := []struct {
			name     string
			priority interface{}
		}{
			{"string", "invalid"},
			{"bool", true},
			{"slice", []int{1, 2, 3}},
			{"map", map[string]int{"priority": 5}},
			{"negative int", -1},
			{"zero", 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := map[string]interface{}{
					"name":     "Test Rule",
					"priority": tc.priority,
					"conditions": map[string]interface{}{
						"all": []interface{}{},
					},
				}

				rule, err := NewRule(options)
				if err == nil {
					t.Errorf("Expected rule creation to fail, but got no error")
				}
				if rule != nil {
					t.Errorf("Expected rule to be nil, but got %v", rule)
				}
			})
		}
	})

	t.Run("Default priority", func(t *testing.T) {
		options := map[string]interface{}{
			"name": "Test Rule",
			"conditions": map[string]interface{}{
				"all": []interface{}{},
			},
		}

		rule, err := NewRule(options)
		if err != nil {
			t.Errorf("Expected rule creation to succeed, but got error: %v", err)
		}
		if rule.Priority != 1 {
			t.Errorf("Expected priority to be 1 (default), but got %d", rule.Priority)
		}
	})
}
