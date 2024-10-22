package rulesengine

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
)

type EventOutcome string

const (
	Success EventOutcome = "success"
	Failure EventOutcome = "failure"
)

// Almanac is a struct that manages fact results lookup and caching within a rules engine.
// It allows storing raw facts, caching results of rules, and logging events (success/failure).
// The Almanac plays a key role in the rules engine by allowing rules to evaluate facts efficiently.
type Almanac struct {
	factMap             FactMap                  // A map storing facts for quick lookup
	allowUndefinedFacts bool                     // Flag to allow or disallow undefined facts
	events              map[EventOutcome][]Event // Maps success or failure outcomes to their events
	ruleResults         []RuleResult             // A slice to store rule evaluation results
	rawFacts            gjson.Result             // The raw input facts in JSON format
	ruleResultsCapacity int                      // Initial capacity for rule results to optimize memory
}

// Options defines the optional settings for the Almanac.
// It includes a flag to allow or disallow the use of undefined facts during rule evaluation.
type Options struct {
	AllowUndefinedFacts *bool // Optional flag to allow undefined facts
}

// NewAlmanac creates and returns a new Almanac instance.
// Params:
// - rf: Raw facts in the form of a gjson.Result.
// - options: Custom settings such as allowing undefined facts.
// - initialCapacity: The initial capacity to allocate for rule results.
// Returns a pointer to a new Almanac.
func NewAlmanac(rf gjson.Result, options Options, initialCapacity int) *Almanac {
	allowUndefinedFacts := false
	if options.AllowUndefinedFacts != nil {
		allowUndefinedFacts = *options.AllowUndefinedFacts
	}

	return &Almanac{
		rawFacts:            rf,
		allowUndefinedFacts: allowUndefinedFacts,
		events:              map[EventOutcome][]Event{"success": {}, "failure": {}},
		ruleResults:         make([]RuleResult, 0, initialCapacity),
		ruleResultsCapacity: initialCapacity,
	}
}

// AddEvent logs an event in the Almanac, marking it as either a success or failure.
// Params:
// - event: The event to be added.
// - outcome: The outcome of the event ("success" or "failure").
// Returns an error if the outcome is invalid.
func (a *Almanac) AddEvent(event Event, outcome EventOutcome) error {
	if outcome != Success && outcome != Failure {
		return errors.New(`outcome required: "success" | "failure"`)
	}
	(a.events)[outcome] = append((a.events)[outcome], event)
	return nil
}

// GetEvents retrieves events logged in the Almanac based on the specified outcome.
// If the outcome is "success" or "failure", it returns the events for that outcome.
// If the outcome is an empty string, it returns all events (success and failure combined).
// Params:
// - outcome: The desired outcome ("success", "failure", or empty string for all events).
// Returns a pointer to a slice of events for the specified outcome.
func (a *Almanac) GetEvents(outcome EventOutcome) *[]Event {
	eventsMap := a.events
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

// AddResult adds a rule evaluation result to the Almanac.
// This function stores the result of a rule once it has been evaluated.
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

func (a *Almanac) AddFact(key string, value *Fact) {
	a.factMap.Set(key, value)
}

// AddRuntimeFact adds a constant fact during runtime
func (a *Almanac) AddRuntimeFact(path string, value ValueNode) error {
	Debug(fmt.Sprintf("almanac::addRuntimeFact id:%s", path))
	f, err := NewFact(path, value, nil)
	if err != nil {
		return err
	}
	a.AddFact(f.Path, f)
	return nil
}

func (a *Almanac) FactValue(path string) (*Fact, error) {
	// Check if the fact is in the cache
	f, ok := a.factMap.Load(path)
	if ok {
		return f, nil
	}

	// If the fact is not in try to read it from the raw facts
	result := a.rawFacts.Get(path)

	if !result.Exists() {
		if a.allowUndefinedFacts {
			return nil, nil
		}
		return nil, fmt.Errorf("undefined fact: %s", path)
	}
	vn := NewValueFromGjson(result)
	// Create a new fact and add it to the cache
	nf, err := NewFact(path, *vn, nil)
	if err != nil {
		return nil, err
	}
	a.AddFact(path, nf)
	return nf, nil
}

func (a *Almanac) GetValue(path string) (interface{}, error) {
	f, err := a.FactValue(path)
	if err != nil || f == nil || f.Value == nil {
		return nil, nil
	}
	switch f.Value.Type {
	case String:
		return f.Value.String, nil
	case Number:
		return f.Value.Number, nil
	case Object:
		return f.Value.Object, nil
	case Array:
		return f.Value.Array, nil
	case Bool:
		return f.Value.Bool, nil
	case Null:
		return nil, nil
	}
	return nil, nil
}
