package rulesengine

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/EventBus"
	"sync"
)

type Event struct {
	Type   string
	Params map[string]interface{}
}

type FactOptions struct {
	Cache    bool
	Priority int
}

type DynamicFactCallback func(params map[string]interface{}, almanac *Almanac) interface{}
type EventCallback func(result *RuleResult) interface{}

type EvaluationResult struct {
	Result             bool        `json:"Result"`
	LeftHandSideValue  interface{} `json:"LeftHandSideValue"`
	RightHandSideValue interface{} `json:"RightHandSideValue"`
	Operator           string      `json:"Operator"`
}

const (
	READY    = "READY"
	RUNNING  = "RUNNING"
	FINISHED = "FINISHED"
)

// RuleProperties represents the properties of a rule.
type RuleProperties struct {
	Conditions TopLevelCondition `json:"conditions"`
	Event      Event             `json:"event"`
	Name       *string           `json:"name,omitempty"`
	Priority   *int              `json:"priority,omitempty"`
	OnSuccess  *EventHandler     `json:"onSuccess,omitempty"`
	OnFailure  *EventHandler     `json:"onFailure,omitempty"`
}

// TopLevelCondition represents the top-level condition, which can be AllConditions, AnyConditions, NotConditions, or ConditionReference.
type TopLevelCondition struct {
	All       *[]ConditionProperties `json:"all,omitempty"`
	Any       *[]ConditionProperties `json:"any,omitempty"`
	Not       *ConditionProperties   `json:"not,omitempty"`
	Condition *string                `json:"condition,omitempty"`
	Name      *string                `json:"name,omitempty"`
	Priority  *int                   `json:"priority,omitempty"`
}

// EventHandler represents an event handler function.
type EventHandler func(event Event, almanac Almanac, ruleResult RuleResult)

// ConditionProperties represents a condition in the rule.
type ConditionProperties struct {
	Fact     string                 `json:"fact"`
	Operator string                 `json:"operator"`
	Value    interface{}            `json:"value"`
	Path     *string                `json:"path,omitempty"`
	Priority *int                   `json:"priority,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Name     *string                `json:"name,omitempty"`
}

func (c *ConditionProperties) SetPriority(priority int) {
	c.Priority = &priority
}

func (c *ConditionProperties) SetName(name string) {
	c.Name = &name
}

type Engine struct {
	Rules                     []*Rule
	AllowUndefinedFacts       bool
	AllowUndefinedConditions  bool
	ReplaceFactsInEventParams bool
	PathResolver              PathResolver
	Operators                 map[string]Operator
	Facts                     sync.Map
	Conditions                sync.Map
	Status                    string
	prioritizedRules          [][]*Rule
	bus                       EventBus.Bus
	mu                        sync.Mutex
}

type RuleEngineOptions struct {
	AllowUndefinedFacts       bool
	AllowUndefinedConditions  bool
	ReplaceFactsInEventParams bool
	PathResolver              PathResolver
}

type RuleConfig struct {
	Name       string      `json:"name"`
	Priority   *int        `json:"priority"`
	Conditions Condition   `json:"conditions"`
	Event      EventConfig `json:"event"`
	OnSuccess  func(result *RuleResult) interface{}
	OnFailure  func(result *RuleResult) interface{}
}

// UnmarshalJSON is a custom JSON unmarshaller for RuleConfig to ensure proper unmarshaling of Condition
func (r *RuleConfig) UnmarshalJSON(data []byte) error {
	// Define an alias to avoid recursion
	type Alias RuleConfig
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	// Unmarshal the data into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Now manually unmarshal and validate the Conditions field
	if err := json.Unmarshal(data, &r.Conditions); err != nil {
		return fmt.Errorf("failed to unmarshal conditions: %v", err)
	}

	return nil
}
