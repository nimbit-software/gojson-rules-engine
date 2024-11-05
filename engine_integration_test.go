package rulesengine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getEngineWithMissingOperator(options *RuleEngineOptions) *Engine {
	engine := NewEngine(nil, options)

	if missingOperator, err := NewOperator("missing", func(a, b *ValueNode) bool {
		return a.Type == Null
	}, nil); err == nil {
		engine.AddOperator(*missingOperator, nil)
	}

	return engine
}

func TestEngineRunUndefinedFacts(t *testing.T) {

	t.Run("should fail when undefined facts are not allowed", func(t *testing.T) {
		rule, _ := NewRule(&RuleConfig{
			Name: "test-rule",
			Conditions: Condition{
				All: []*Condition{
					{
						Fact:     "some.undefinedFact",
						Operator: "missing",
						Value:    ValueNode{Type: String, String: "someValue"},
					},
				},
			},
			Event: EventConfig{
				Type: "test-event",
			},
		})

		engine := getEngineWithMissingOperator(&RuleEngineOptions{
			AllowUndefinedFacts: false,
		})
		engine.AddRule(rule)

		_, err := engine.Run(context.Background(), []byte(`{}`))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undefined fact: some.undefinedFact")
	})

	t.Run("should succeed when undefined facts are allowed", func(t *testing.T) {
		rule, _ := NewRule(&RuleConfig{
			Name: "test-rule",
			Conditions: Condition{
				All: []*Condition{
					{
						Fact:     "some.undefinedFact",
						Operator: "missing",
					},
				},
			},
			Event: EventConfig{
				Type: "test-event",
			},
		})

		engine := getEngineWithMissingOperator(&RuleEngineOptions{
			AllowUndefinedFacts: true,
		})
		engine.AddRule(rule)
		result, err := engine.Run(context.Background(), []byte(`{ some: {} }`))

		assert.NoError(t, err)
		assert.NotNil(t, result)

		events := result["events"].(*[]Event)
		assert.Len(t, *events, 1)
		assert.Equal(t, "test-event", (*events)[0].Type)
	})
}
