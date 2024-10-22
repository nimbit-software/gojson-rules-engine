package rulesengine

import (
	"encoding/json"
	"sync"
)

// RuleResult represents the result of a rule evaluation
type RuleResult struct {
	Conditions Condition
	Event      Event
	Priority   int
	Name       string
	Result     *bool
	mu         sync.Mutex
}

// NewRuleResult creates a new RuleResult instance
func NewRuleResult(conditions Condition, event Event, priority int, name string) *RuleResult {
	return &RuleResult{
		Conditions: conditions,
		Event:      event,
		Priority:   priority,
		Name:       name,
		Result:     nil,
	}
}

// SetResult sets the result of the rule evaluation
func (rr *RuleResult) SetResult(result *bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.Result = result
}

// ResolveEventParams resolves the event parameters using the given almanac
func (rr *RuleResult) ResolveEventParams(almanac *Almanac) error {
	if IsObjectLike(rr.Event.Params) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errorsCh := make(chan error, len(rr.Event.Params))

		for key, value := range rr.Event.Params {
			wg.Add(1)
			go func(key string, value interface{}) {
				defer wg.Done()
				// check if value is a string
				if IsObjectLike(value) {
					valMap, ok := value.(map[string]interface{})
					if ok {
						if factPath, ok := valMap["fact"].(string); ok {
							resolvedValue, err := almanac.GetValue(factPath)
							if err != nil {
								errorsCh <- err
								return
							}

							mu.Lock()
							rr.Event.Params[key] = resolvedValue
							mu.Unlock()
						}
					}
				}

			}(key, value)
		}

		wg.Wait()
		close(errorsCh)

		if len(errorsCh) > 0 {
			return <-errorsCh
		}
	}
	return nil
}

// ToJSON converts the rule result to a JSON-friendly structure
func (rr *RuleResult) ToJSON(stringify bool) (interface{}, error) {
	props := map[string]interface{}{
		"conditions": rr.Conditions,
		"event":      rr.Event,
		"priority":   rr.Priority,
		"name":       rr.Name,
		"result":     rr.Result,
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
