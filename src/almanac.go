package src

import (
	"errors"
	"fmt"
	"github.com/oliveagle/jsonpath"
	"sync"
)

// PathResolver is a shared_types alias for a function that resolves a path within a JSON-like structure
type PathResolver func(value interface{}, path string) (interface{}, error)

// DefaultPathResolver is the default function to resolve a path within a JSON-like structure
func DefaultPathResolver(value interface{}, path string) (interface{}, error) {
	res, err := jsonpath.JsonPathLookup(value, path)
	return res, err
}

// Almanac represents fact results lookup and caching
type Almanac struct {
	factMap             sync.Map
	factResultsCache    sync.Map
	allowUndefinedFacts bool
	pathResolver        PathResolver
	events              *map[string][]interface{} // TODO USE REAL TYPE
	ruleResults         *[]RuleResult
}

type Options struct {
	AllowUndefinedFacts *bool
	PathResolver        *PathResolver
}

// NewAlmanac creates a new Almanac instance
func NewAlmanac(options Options) *Almanac {
	allowUndefinedFacts := false
	if options.AllowUndefinedFacts != nil {
		allowUndefinedFacts = *options.AllowUndefinedFacts
	}
	pathResolver := DefaultPathResolver
	if *(options.PathResolver) != nil {
		pathResolver = *options.PathResolver
	}

	return &Almanac{
		allowUndefinedFacts: allowUndefinedFacts,
		pathResolver:        pathResolver,
		events:              &map[string][]interface{}{"success": {}, "failure": {}},
		ruleResults:         &[]RuleResult{},
	}
}

// AddEvent adds a success or failure event
func (a *Almanac) AddEvent(event interface{}, outcome string) error {
	if outcome != "success" && outcome != "failure" {
		return errors.New(`outcome required: "success" | "failure"`)
	}
	(*a.events)[outcome] = append((*a.events)[outcome], event)
	return nil
}

// GetEvents retrieves events based on the outcome
func (a *Almanac) GetEvents(outcome string) *[]interface{} {
	eventsMap := *a.events // Dereference the pointer to access the map
	if outcome != "" {
		// Return a pointer to the slice for the specified outcome
		events, exists := eventsMap[outcome]
		if exists {
			return &events
		}
		// Return nil or an empty slice pointer if the outcome does not exist
		return &[]interface{}{}
	}

	// Combine "success" and "failure" slices if outcome is an empty string
	combinedEvents := append(eventsMap["success"], eventsMap["failure"]...)
	return &combinedEvents
}

// AddResult adds a rule result
func (a *Almanac) AddResult(ruleResult *RuleResult) {
	*a.ruleResults = append(*a.ruleResults, *ruleResult)
}

// GetResults retrieves all rule results
func (a *Almanac) GetResults() *[]RuleResult {
	return a.ruleResults
}

// getFact retrieves a fact by its ID
func (a *Almanac) getFact(factId string) (*Fact, error) {
	value, ok := a.factMap.Load(factId)
	if !ok {
		return nil, fmt.Errorf("undefined fact: %s", factId)
	}
	f, ok := value.(*Fact)
	if !ok {
		return nil, fmt.Errorf("invalid fact shared_types for fact: %s", factId)
	}
	return f, nil
}

// addConstantFact adds a constant fact
func (a *Almanac) addConstantFact(f *Fact) {
	a.factMap.Store(f.ID, f)
	a.setFactValue(f, map[string]interface{}{}, f.Value)
}

// setFactValue sets the computed value of a fact
func (a *Almanac) setFactValue(f *Fact, params map[string]interface{}, value interface{}) {
	cacheKey, _ := f.GetCacheKey(params)
	factValue := value
	if cacheKey != 0 {
		a.factResultsCache.Store(cacheKey, factValue)
	}
}

// AddFact adds a fact definition to the engine
func (a *Almanac) AddFact(id interface{}, valueOrMethod interface{}, options *FactOptions) *Almanac {
	var factId string
	var f *Fact
	switch v := id.(type) {
	case *Fact:
		factId = v.ID
		f = v
	case string:
		factId = v
		f, _ = NewFact(factId, valueOrMethod, options)
	default:
		Debug("invalid shared_types for id")
		return a
	}
	Debug(fmt.Sprintf("almanac::addFact id:%s", factId))
	a.factMap.Store(factId, f)
	if f.IsConstant() {
		a.setFactValue(f, map[string]interface{}{}, f.Value)
	}
	return a
}

// AddRuntimeFact adds a constant fact during runtime
func (a *Almanac) AddRuntimeFact(factId string, value interface{}) {
	Debug(fmt.Sprintf("almanac::addRuntimeFact id:%s", factId))
	f, _ := NewFact(factId, value, nil)
	a.addConstantFact(f)
}

// FactValue returns the value of a fact
func (a *Almanac) FactValue(factId string, params map[string]interface{}, path string) (interface{}, error) {
	f, err := a.getFact(factId)
	if err != nil {
		if a.allowUndefinedFacts {
			return nil, nil
		}
		return nil, &UndefinedFactError{Message: fmt.Sprintf("Undefined fact: %s", factId)}
	}

	var factValue interface{}
	if f.IsConstant() {
		factValue = f.Calculate(params, a)
	} else {
		cacheKey, _ := f.GetCacheKey(params)
		if cacheVal, ok := a.factResultsCache.Load(cacheKey); ok {
			factValue = cacheVal
			Debug(fmt.Sprintf("almanac::factValue cache hit for fact:%s", factId))
		} else {
			Debug(fmt.Sprintf("almanac::factValue cache miss for fact:%s; calculating", factId))
			factValue = f.Calculate(params, a)
			a.setFactValue(f, params, factValue)
		}
	}

	if path != "" {
		Debug(fmt.Sprintf("condition::evaluate extracting object property %s", path))
		if IsObjectLike(factValue) {
			pathValue, err := a.pathResolver(factValue, path)
			if err != nil {
				return nil, err
			}
			Debug(fmt.Sprintf("condition::evaluate extracting object property %s, received: %v", path, pathValue))
			return pathValue, nil
		}
		Debug(fmt.Sprintf("condition::evaluate could not compute object path(%s) of non-object: %v <%T>; continuing with %v", path, factValue, factValue, factValue))
	}
	return factValue, nil
}

// GetValue interprets value as either a primitive or a fact
func (a *Almanac) GetValue(value interface{}) (interface{}, error) {
	if IsObjectLike(value) {
		valMap, ok := value.(map[string]interface{})
		if ok {
			if factId, ok := valMap["fact"].(string); ok {
				return a.FactValue(factId, valMap["params"].(map[string]interface{}), valMap["path"].(string))
			}
		}
	}
	return value, nil
}
