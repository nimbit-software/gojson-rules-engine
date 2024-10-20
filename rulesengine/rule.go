package rulesengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/asaskevich/EventBus"
)

// Rule represents a rule in the rule engine
type Rule struct {
	Priority   int
	Name       string
	Conditions Condition
	RuleEvent  Event
	Engine     *Engine
	bus        EventBus.Bus
	mu         sync.Mutex
}

type EventConfig struct {
	Type   string
	Params *map[string]interface{}
}

type RuleConfig struct {
	Name       string      `json:"name"`
	Priority   *int        `json:"priority"`
	Conditions Condition   `json:"conditions"`
	Event      EventConfig `json:"event"`
	OnSuccess  func(result *RuleResult) interface{}
	OnFailure  func(result *RuleResult) interface{}
}

// NewRule creates a new Rule instance
func NewRule(config *RuleConfig) (*Rule, error) {
	rule := &Rule{
		Name:       config.Name,
		Priority:   1,
		Conditions: config.Conditions,
		RuleEvent: Event{
			Type: "unknown",
		},
		bus: EventBus.New(),
	}
	// RULE PRIORITY
	if config.Priority != nil {
		rule.setName(config.Priority)
	}

	// Subscribe the onSuccess callback if it exists
	if config.OnSuccess != nil {
		err := rule.bus.Subscribe("success", config.OnSuccess)
		if err != nil {
			return nil, err
		}
	}

	// Subscribe the onFailure callback if it exists
	if config.OnFailure != nil {
		err := rule.bus.Subscribe("failure", config.OnFailure)
		if err != nil {
			return nil, err
		}
	}

	// Set the event if the type is provided
	if config.Event.Type != "" {
		rule.setEvent(config.Event)
	} else {
		return nil, errors.New("event type is required")
	}

	return rule, nil
}

// SetPriority sets the priority of the rule
func (r *Rule) setPriority(priority int) {
	if priority <= 0 {
		panic("Priority must be greater than zero")
	}
	r.Priority = priority
}

// SetName sets the name of the rule
func (r *Rule) setName(name interface{}) {
	if name == nil {
		panic("Rule 'name' must be defined")
	}
	r.Name = fmt.Sprintf("%v", name)
}

func (r *Rule) GetName() string {
	return r.Name
}

//// SetConditions sets the conditions to run when evaluating the rule
//func (r *Rule) setConditions(conditions map[string]interface{}) {
//	if _, ok := conditions["all"]; !ok {
//		if _, ok := conditions["any"]; !ok {
//			if _, ok := conditions["not"]; !ok {
//				if _, ok := conditions["condition"]; !ok {
//					panic(`"conditions" root must contain a single instance of "all", "any", "not", or "condition"`)
//				}
//			}
//		}
//	}
//	r.Conditions, _ = NewCondition(conditions)
//}

// SetEvent sets the event to emit when the conditions evaluate truthy
func (r *Rule) setEvent(event EventConfig) {
	r.RuleEvent = Event{
		Type: event.Type,
	}
	if event.Params != nil {
		r.RuleEvent.Params = *event.Params
	}
}

// GetEvent returns the event object
func (r *Rule) GetEvent() Event {
	return r.RuleEvent
}

// GetPriority returns the priority
func (r *Rule) GetPriority() int {
	return r.Priority
}

// GetConditions returns the event object
func (r *Rule) GetConditions() *Condition {
	return &r.Conditions
}

// GetEngine returns the engine object
func (r *Rule) GetEngine() *Engine {
	return r.Engine
}

// SetEngine sets the engine to run the rules under
func (r *Rule) SetEngine(engine *Engine) {
	r.Engine = engine
}

// ToJSON converts the rule to a JSON-friendly structure
func (r *Rule) ToJSON(stringify bool) (interface{}, error) {
	conditions, err := r.Conditions.ToJSON(false)
	if err != nil {
		return nil, err
	}

	props := map[string]interface{}{
		"conditions": conditions,
		"priority":   r.Priority,
		"event":      r.RuleEvent,
		"name":       r.Name,
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

// Evaluate evaluates the rule
func (r *Rule) Evaluate(ctx *ExecutionContext, almanac *Almanac) (*RuleResult, error) {
	ruleResult := NewRuleResult(r.Conditions, r.RuleEvent, r.Priority, r.Name)

	var realize func(*Condition) (bool, error)
	var evaluateCondition func(ctx *ExecutionContext, cond *Condition) (bool, error)
	var prioritizeAndRun func(ctx *ExecutionContext, cond []*Condition, operator string) (bool, error)

	realize = func(conditionReference *Condition) (bool, error) {
		cond, ok := r.Engine.Conditions.Load(conditionReference.Condition)
		if !ok {
			if r.Engine.AllowUndefinedConditions {
				conditionReference.Result = false
				return false, nil
			}
			return false, fmt.Errorf("no condition %s exists", conditionReference.Condition)
		}
		conditionReference.Condition = ""

		err := DeepCopy(&cond, &conditionReference)
		if err != nil {
			return false, err
		}
		return evaluateCondition(ctx, conditionReference)
	}

	evaluateCondition = func(ctx *ExecutionContext, cond *Condition) (bool, error) {
		if cond.IsConditionReference() {
			// If this is a condition reference, realize it before evaluation
			return realize(cond)
		}

		var result bool
		var err error

		// Evaluate 'all' block if it exists
		if cond.All != nil && len(cond.All) > 0 {
			result, err = prioritizeAndRun(ctx, cond.All, "all")
			if err != nil || !result {
				// Early exit if 'all' block fails
				ctx.StopEarly = true
				ctx.Message = "Stopping early due to 'all' condition failure"
				ctx.Cancel()
				return result, err
			}
		}

		// Evaluate 'any' block if it exists
		if cond.Any != nil && len(cond.Any) > 0 {
			result, err = prioritizeAndRun(ctx, cond.Any, "any")
			if err != nil {
				return false, err
			}
			if result {
				// Early exit if 'any' block succeeds
				ctx.StopEarly = true
				ctx.Message = "Stopping early due to 'any' condition success"
				ctx.Cancel()
				return result, nil
			}
		}

		// Evaluate 'not' block if it exists
		if cond.Not != nil {
			result, err = prioritizeAndRun(ctx, []*Condition{cond.Not}, "not")
			if err != nil {
				return false, err
			}
			if !result {
				// If 'not' block is false, return true (because it's negation)
				return true, nil
			} else {
				// If 'not' block is true, return false
				return false, nil
			}
		}

		// Base case: If there's no 'any', 'all', or 'not', it's a simple condition
		if !cond.IsBooleanOperator() {
			evaluationResult, err := cond.Evaluate(almanac, r.Engine.Operators)
			if err != nil {
				return false, err
			}
			cond.FactResult = evaluationResult.LeftHandSideValue
			cond.Result = evaluationResult.Result
			return evaluationResult.Result, nil
		}

		// Default to false if none of the above cases match
		return result, err
	}

	evaluateConditions := func(ctx *ExecutionContext, conditions []*Condition, method func([]bool) bool) (bool, error) {
		if len(conditions) == 0 {
			return true, nil
		}

		results := make([]bool, len(conditions))
		errs := make(chan error, len(conditions))
		resCh := make(chan struct {
			index  int
			result bool
		}, len(conditions))

		var wg sync.WaitGroup
		wg.Add(len(conditions))

		for i, cond := range conditions {
			go func(i int, cond *Condition) {
				defer wg.Done()

				select {
				case <-ctx.Done():
					Debug("Context cancelled in evaluateConditions goroutine")
					return
				default:
					result, err := evaluateCondition(ctx, cond)
					if err != nil {
						errs <- err
						return
					}
					resCh <- struct {
						index  int
						result bool
					}{index: i, result: result}
				}
			}(i, cond)
		}

		// Close channels once all goroutines are done
		go func() {
			wg.Wait()
			close(errs)
			close(resCh)
		}()

		// Collect results
		for i := 0; i < len(conditions); i++ {
			select {
			case err := <-errs:
				return false, err
			case res := <-resCh:
				results[res.index] = res.result

				if ctx.StopEarly {
					return method(results), nil
				}
			}
		}

		return method(results), nil
	}

	prioritizeAndRun = func(ctx *ExecutionContext, conditions []*Condition, operator string) (bool, error) {
		if len(conditions) == 0 {
			return true, nil
		}
		if len(conditions) == 1 {
			return evaluateCondition(ctx, conditions[0])
		}

		var method func([]bool) bool
		switch operator {
		case "all":
			method = func(results []bool) bool {
				for _, result := range results {
					if !result {
						return false
					}
				}
				return true
			}
		case "any":
			method = func(results []bool) bool {
				for _, result := range results {
					if result {
						return true
					}
				}
				return false
			}
		case "not":
			method = func(results []bool) bool {
				return !results[0]
			}
		default:
			return false, errors.New("invalid operator")
		}

		// Prioritize conditions based on priority
		orderedSets := r.prioritizeConditions(conditions)
		for _, set := range orderedSets {
			if ctx.StopEarly {
				return false, nil
			}
			result, err := evaluateConditions(ctx, set, method)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// Main evaluation logic
	processResult := func(result bool) (*RuleResult, error) {
		ruleResult.SetResult(&result)
		if r.Engine.ReplaceFactsInEventParams {
			if err := ruleResult.ResolveEventParams(almanac); err != nil {
				return nil, err
			}
		}
		event := "failure"
		if result {
			event = "success"
		}
		go r.bus.Publish(event, ruleResult)
		return ruleResult, nil
	}

	var result bool
	var err error

	conditions := map[string][]*Condition{}

	if ruleResult.Conditions.Any != nil && len(ruleResult.Conditions.Any) > 0 {
		conditions["any"] = ruleResult.Conditions.Any
	}

	if ruleResult.Conditions.All != nil && len(ruleResult.Conditions.All) > 0 {
		conditions["all"] = ruleResult.Conditions.All
	}

	if ruleResult.Conditions.Not != nil {
		conditions["not"] = []*Condition{ruleResult.Conditions.Not} // Wrap `Not` in a slice
	}

	// If no conditions are provided, realize the default conditions
	if ruleResult.Conditions.All == nil && ruleResult.Conditions.Any == nil && ruleResult.Conditions.Not == nil {
		result, err = realize(&r.Conditions)
		if err != nil {
			return nil, err
		}
	} else {
		// Iterate over the conditions and execute prioritizeAndRun if the condition is present
		for operator, condition := range conditions {
			result, err = prioritizeAndRun(ctx, condition, operator)
			if err != nil {
				return nil, err
			}
		}
	}

	return processResult(result)
}

func (r *Rule) prioritizeConditions(conditions []*Condition) [][]*Condition {
	factSets := make(map[int][]*Condition)
	for _, cond := range conditions {
		priority := cond.Priority
		if priority == 0 {
			f, _ := r.Engine.Facts.Load(cond.Fact)
			if f != nil {
				priority = f.(*Fact).Priority
			}
		}
		factSets[priority] = append(factSets[priority], cond)
	}

	var keys []int
	for k := range factSets {
		keys = append(keys, k)
	}

	// Sort keys in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	var result [][]*Condition
	for _, k := range keys {
		result = append(result, factSets[k])
	}
	return result
}
