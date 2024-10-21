package rulesengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"sort"
	"sync"

	"github.com/asaskevich/EventBus"
)

func DefaultRuleEngineOptions() *RuleEngineOptions {
	return &RuleEngineOptions{
		AllowUndefinedFacts:       false,
		AllowUndefinedConditions:  false,
		ReplaceFactsInEventParams: false,
	}
}

// NewEngine creates a new Engine instance
func NewEngine(rules []*Rule, options *RuleEngineOptions) *Engine {
	if options == nil {
		options = DefaultRuleEngineOptions()
	}

	engine := &Engine{
		Rules:                     []*Rule{},
		Operators:                 make(map[string]Operator),
		Status:                    READY,
		bus:                       EventBus.New(),
		AllowUndefinedConditions:  options.AllowUndefinedConditions,
		AllowUndefinedFacts:       options.AllowUndefinedFacts,
		ReplaceFactsInEventParams: options.ReplaceFactsInEventParams,
	}

	for _, r := range rules {
		err := engine.AddRule(r)
		if err != nil {
			return nil
		}
	}
	for _, o := range DefaultOperators() {
		engine.AddOperator(o, nil)
	}
	return engine
}

func (e *Engine) AddRule(rule *Rule) error {
	if rule == nil {
		return errors.New("engine: rule is required")
	}

	rule.SetEngine(e)
	e.Rules = append(e.Rules, rule)
	e.prioritizedRules = nil
	return nil
}

func (e *Engine) AddRuleFromMap(rp *RuleConfig) error {
	if rp == nil {
		return errors.New("engine: AddRuleFromMap invalid configuration")
	}

	r, _ := NewRule(rp)
	r.SetEngine(e)
	e.Rules = append(e.Rules, r)
	e.prioritizedRules = nil
	return nil
}

func (e *Engine) AddRules(rules []*Rule) error {
	for _, r := range rules {
		err := e.AddRule(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateRule updates a rule inEvaluator the engine
func (e *Engine) UpdateRule(r *Rule) error {
	ruleIndex := -1
	for i, ruleInEngine := range e.Rules {
		if ruleInEngine.Name == r.Name {
			ruleIndex = i
			break
		}
	}

	if ruleIndex > -1 {
		e.Rules = append(e.Rules[:ruleIndex], e.Rules[ruleIndex+1:]...)
		err := e.AddRule(r)
		if err != nil {
			return err
		}
		e.prioritizedRules = nil
		return nil
	}
	return errors.New("engine: updateRule() rule not found")
}

func (e *Engine) RemoveRule(rule *Rule) bool {
	index := -1
	for i, r := range e.Rules {
		if r == rule {
			index = i
			break
		}
	}

	if index > -1 {
		e.Rules = append(e.Rules[:index], e.Rules[index+1:]...)
		e.prioritizedRules = nil // reset prioritized rules
		return true
	}
	return false
}

func (e *Engine) RemoveRuleByName(name string) bool {
	var filteredRules []*Rule
	for _, r := range e.Rules {
		if r.Name != name {
			filteredRules = append(filteredRules, r)
		}
	}

	if len(filteredRules) != len(e.Rules) {
		e.Rules = filteredRules
		e.prioritizedRules = nil // reset prioritized rules
		return true
	}
	return false
}

func (e *Engine) GetRules() []*Rule {
	return e.Rules
}

// RemoveCondition removes a condition that has previously been added to this engine
func (e *Engine) RemoveCondition(name string) bool {
	_, ok := e.Conditions.Load(name)
	if ok {
		e.Conditions.Delete(name)
	}
	return ok
}

// AddOperator adds a custom operator definition
func (e *Engine) AddOperator(operatorOrName interface{}, cb func(gjson.Result, gjson.Result) bool) {
	var op Operator
	switch v := operatorOrName.(type) {
	case Operator:
		op = v
	case string:
		newOpp, _ := NewOperator(v, cb, nil)
		op = *newOpp
	}
	Debug(fmt.Sprintf("engine::addOperator name:%s", op.Name))
	e.Operators[op.Name] = op
}

// RemoveOperator removes a custom operator definition
func (e *Engine) RemoveOperator(operatorOrName interface{}) bool {
	var operatorName string
	switch v := operatorOrName.(type) {
	case Operator:
		operatorName = v.Name
	case string:
		operatorName = v
	}
	_, ok := e.Operators[operatorName]
	if ok {
		delete(e.Operators, operatorName)
	}
	return ok
}

//// AddFact adds a fact definition to the engine
//func (e *Engine) AddFact(id interface{}, valueOrMethod interface{}, options *FactOptions) *Engine {
//	var factId string
//	var f *Fact
//	switch v := id.(type) {
//	case *Fact:
//		factId = v.ID
//		f = v
//	case string:
//		factId = v
//		f, _ = NewFact(factId, valueOrMethod, options)
//	}
//	Debug(fmt.Sprintf("engine::addFact id:%s", factId))
//	e.Facts.Store(factId, f)
//	return e
//}
//
//// RemoveFact removes a fact definition from the engine
//func (e *Engine) RemoveFact(factOrId interface{}) bool {
//	var factId string
//	switch v := factOrId.(type) {
//	case *Fact:
//		factId = v.ID
//	case string:
//		factId = v
//	}
//	_, ok := e.Facts.Load(factId)
//	if ok {
//		e.Facts.Delete(factId)
//	}
//	return ok
//}

// PrioritizeRules iterates over the engine rules, organizing them by highest -> lowest priority
func (e *Engine) PrioritizeRules() [][]*Rule {
	if e.prioritizedRules == nil {
		ruleSets := make(map[int][]*Rule)
		for _, r := range e.Rules {
			priority := r.GetPriority()
			ruleSets[priority] = append(ruleSets[priority], r)
		}

		var keys []int
		for k := range ruleSets {
			keys = append(keys, k)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(keys)))

		for _, k := range keys {
			e.prioritizedRules = append(e.prioritizedRules, ruleSets[k])
		}
	}
	return e.prioritizedRules
}

// Stop stops the rules engine from running the next priority set of Rules
func (e *Engine) Stop() *Engine {
	e.Status = FINISHED
	return e
}

//// GetFact returns a fact by fact-id
//func (e *Engine) GetFact(factId string) *Fact {
//	f, _ := e.Facts.Load(factId)
//	if f == nil {
//		return nil
//	}
//	return f.(*Fact)
//}

// EvaluateRules runs an array of rules
func (e *Engine) EvaluateRules(rules []*Rule, almanac *Almanac, ctx *ExecutionContext) error {
	// CHECK STATE OF ENGINE
	if e.Status != RUNNING {
		Debug(fmt.Sprintf("engine::run status:%s; skipping remaining rules", e.Status))
		return nil
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(rules))
	results := make(chan *RuleResult, len(rules))

	for _, r := range rules {
		if ctx.StopEarly {
			break
		}

		wg.Add(1)
		go func(rule *Rule) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				Debug("Context cancelled inEvaluator goroutine")
				return
			default:
				ruleResult, err := rule.Evaluate(ctx, almanac)
				if err != nil {
					errs <- err
					return
				}

				Debug(fmt.Sprintf("engine::run ruleResult:%v", ruleResult.Result))
				results <- ruleResult
				Debug("Result sent to results channel inEvaluator goroutine")
			}
		}(r)
	}

	// Close results and errors channels after all goroutines complete
	go func() {
		wg.Wait()
		Debug("All goroutines completed")
		close(results)
		close(errs)
	}()

	// Collect results
	for ruleResult := range results {
		Debug("Received result from results channel")
		almanac.AddResult(ruleResult)
		if ruleResult.Result != nil && *ruleResult.Result {
			err := almanac.AddEvent(ruleResult.Event, "success")
			if err != nil {
				Debug(fmt.Sprintf("Error adding success event: %v", err))
				return err
			}
			e.bus.Publish("success", ruleResult.Event, almanac, ruleResult)
			e.bus.Publish(ruleResult.Event.Type, ruleResult.Event.Params, almanac, ruleResult)
		} else {
			err := almanac.AddEvent(ruleResult.Event, "failure")
			if err != nil {
				Debug(fmt.Sprintf("Error adding failure event: %v", err))
				return err
			}
			e.bus.Publish("failure", ruleResult.Event, almanac, ruleResult)
		}
	}

	// Check for errors
	for err := range errs {
		Debug("Received error from errs channel")
		return err
	}

	return nil
}

func (e *Engine) Run(ctx context.Context, input []byte) (map[string]interface{}, error) {
	return e.runInternal(ctx, input)
}

func (e *Engine) RunWithMap(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	factBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshaling input map: %v", err)
	}
	return e.runInternal(ctx, factBytes)
}

// Run runs the rules engine
func (e *Engine) runInternal(ctx context.Context, facts []byte) (map[string]interface{}, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("engine::run recovered from panic: %v", r)
		}
	}()

	Debug("engine::run started")
	e.Status = RUNNING

	parsedFacts := gjson.ParseBytes(facts)

	almanacInstance := NewAlmanac(parsedFacts, Options{
		AllowUndefinedFacts: &e.AllowUndefinedFacts,
	}, len(e.Rules))

	e.Facts.Range(func(_, value interface{}) bool {
		//f := value.(*Fact)
		//almanacInstance.AddFact(f, nil, nil)
		return true
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Run Context
	execCtx := &ExecutionContext{
		Context: ctx,
		Cancel:  cancel,
	}

	orderedSets := e.PrioritizeRules()
	for _, set := range orderedSets {
		if err := e.EvaluateRules(set, almanacInstance, execCtx); err != nil {
			return nil, err
		}
		if execCtx.StopEarly {
			break
		}
	}

	e.Status = FINISHED
	Debug("engine::run completed")

	ruleResults := almanacInstance.GetResults()
	var results []*RuleResult
	var failureResults []*RuleResult

	// Safely dereference ruleResults before iterating
	if ruleResults != nil {
		for _, ruleResult := range ruleResults {
			// Safely check if ruleResult.Result is not nil and true
			if ruleResult.Result != nil && *ruleResult.Result {
				results = append(results, &ruleResult)
			} else {
				failureResults = append(failureResults, &ruleResult)
			}
		}
	}

	return map[string]interface{}{
		"almanac":        almanacInstance,
		"results":        results,
		"failureResults": failureResults,
		"events":         almanacInstance.GetEvents("success"),
		"failureEvents":  almanacInstance.GetEvents("failure"),
	}, err
}
