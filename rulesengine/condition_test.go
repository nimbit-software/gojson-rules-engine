package rulesengine

import (
	"testing"
)

func TestNewCondition(t *testing.T) {
	t.Run("Valid priority types", func(t *testing.T) {
		testCases := []struct {
			name     string
			priority interface{}
			expected int
		}{
			{"float64", float64(2.0), 2},
			{"float32", float32(2.0), 2},
			{"int64", int64(3), 3},
			{"int", 4, 4},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				conditions := map[string]interface{}{
					"all": []interface{}{
						map[string]interface{}{
							"fact":     "age",
							"operator": "greaterThan",
							"value":    18,
							"priority": tc.priority,
						},
					},
				}

				condition, err := NewCondition(conditions)
				if err != nil {
					t.Errorf("Expected condition creation to succeed, but got error: %v", err)
				}
				if condition.All[0].Priority != tc.expected {
					t.Errorf("Expected priority to be %d, but got %d", tc.expected, condition.All[0].Priority)
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
				conditions := map[string]interface{}{
					"all": []interface{}{
						map[string]interface{}{
							"fact":     "age",
							"operator": "greaterThan",
							"value":    18,
							"priority": tc.priority,
						},
					},
				}

				condition, err := NewCondition(conditions)
				if err == nil {
					t.Errorf("Expected condition creation to fail, but got no error")
				}
				if condition != nil {
					t.Errorf("Expected condition to be nil, but got %v", condition)
				}
			})
		}
	})

	t.Run("Default priority", func(t *testing.T) {
		conditions := map[string]interface{}{
			"all": []interface{}{
				map[string]interface{}{
					"fact":     "age",
					"operator": "greaterThan",
					"value":    18,
				},
			},
		}

		condition, err := NewCondition(conditions)
		if err != nil {
			t.Errorf("Expected condition creation to succeed, but got error: %v", err)
		}
		if condition.All[0].Priority != 0 {
			t.Errorf("Expected priority to be 0 (default), but got %d", condition.All[0].Priority)
		}
	})
}
