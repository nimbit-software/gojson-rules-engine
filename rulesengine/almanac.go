package rulesengine

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"sync"
)

type EventOutcome string

const (
	Success EventOutcome = "success"
	Failure EventOutcome = "failure"
)

// Almanac represents fact results lookup and caching
type Almanac struct {
	factMap             sync.Map
	factResultsCache    sync.Map
	allowUndefinedFacts bool
	events              map[EventOutcome][]Event
	ruleResults         []RuleResult
	facts               gjson.Result
	ruleResultsCapacity int
}

type Options struct {
	AllowUndefinedFacts *bool
}

// NewAlmanac creates a new Almanac instance
func NewAlmanac(facts gjson.Result, options Options, initialCapacity int) *Almanac {
	allowUndefinedFacts := false
	if options.AllowUndefinedFacts != nil {
		allowUndefinedFacts = *options.AllowUndefinedFacts
	}

	return &Almanac{
		facts:               facts,
		allowUndefinedFacts: allowUndefinedFacts,
		events:              map[EventOutcome][]Event{"success": {}, "failure": {}},
		ruleResults:         make([]RuleResult, 0, initialCapacity),
		ruleResultsCapacity: initialCapacity,
	}
}

// AddEvent adds a success or failure event
func (a *Almanac) AddEvent(event Event, outcome EventOutcome) error {
	if outcome != Success && outcome != Failure {
		return errors.New(`outcome required: "success" | "failure"`)
	}
	(a.events)[outcome] = append((a.events)[outcome], event)
	return nil
}

// GetEvents retrieves events based on the outcome
func (a *Almanac) GetEvents(outcome EventOutcome) *[]Event {
	eventsMap := a.events // Dereference the pointer to access the map
	if outcome != "" {
		// Return a pointer to the slice for the specified outcome
		events, exists := eventsMap[outcome]
		if exists {
			return &events
		}
		// Return nil or an empty slice pointer if the outcome does not exist
		return &[]Event{}
	}

	// Combine "success" and "failure" slices if outcome is an empty string
	combinedEvents := append(eventsMap["success"], eventsMap["failure"]...)
	return &combinedEvents
}

// AddResult adds a rule result
func (a *Almanac) AddResult(ruleResult *RuleResult) {
	if len(a.ruleResults) == a.ruleResultsCapacity {
		// Double the capacity when we need to grow
		newCapacity := a.ruleResultsCapacity * 2
		if newCapacity == 0 {
			newCapacity = 4 // Start with a small capacity if it was initially 0
		}
		newSlice := make([]RuleResult, len(a.ruleResults), newCapacity)
		copy(newSlice, a.ruleResults)
		a.ruleResults = newSlice
		a.ruleResultsCapacity = newCapacity
	}
	a.ruleResults = append(a.ruleResults, *ruleResult)
}

// GetResults retrieves all rule results
func (a *Almanac) GetResults() []RuleResult {
	return a.ruleResults
}

// getFact retrieves a fact by its ID
func (a *Almanac) getFact(factId string) (*gjson.Result, error) {
	value, ok := a.factMap.Load(factId)
	if !ok {
		return nil, fmt.Errorf("undefined fact: %s", factId)
	}
	f, ok := value.(*gjson.Result)
	if !ok {
		return nil, fmt.Errorf("invalid fact shared_types for fact: %s", factId)
	}
	return f, nil
}

// AddRuntimeFact adds a constant fact during runtime
func (a *Almanac) AddRuntimeFact(factId string, value interface{}) error {
	Debug(fmt.Sprintf("almanac::addRuntimeFact id:%s", factId))
	str, err := sjson.Set(a.facts.String(), factId, value)
	if err != nil {
		return err
	}
	a.facts = gjson.Parse(str)
	return nil
}

func (a *Almanac) FactValue(path string) (gjson.Result, error) {
	result := a.facts.Get(path)

	if !result.Exists() {
		if a.allowUndefinedFacts {
			return result, nil
		}
		return result, fmt.Errorf("undefined fact: %s", path)
	}
	return result, nil
}

func (a *Almanac) GetValue(path string) (interface{}, error) {
	result := a.facts.Get(path)
	switch result.Type {
	case gjson.String:
		return result.String(), nil
	case gjson.Number:
		return result.Num, nil
	case gjson.JSON:
		return result.Value(), nil
	case gjson.True:
		return true, nil
	case gjson.False:
		return false, nil
	case gjson.Null:
		return nil, nil
	}

	return nil, nil
}
