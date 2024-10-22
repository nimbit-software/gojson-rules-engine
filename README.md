![json-rules-engine](./gopher.png)

# GO-(json-rules-engine)

This projects strives to be a drop in replacement for the original [json-rules-engine](https://github.com/CacheControl/json-rules-engine/) project, but written in Go. A big thanks on the team from cache-control for the current project
With the goal of offering better performance and through go's concurrency model. 

## Project Status
The project is still in its initial stages and is not yet ready for production use, but we are working to get it there since we want to use it in production.


A rules engine expressed in JSON

* [Synopsis](#synopsis)
* [Features](#features)
* [Installation](#installation)
* [Todos](#todos)
* [Operators](#operators)
* [Examples](#examples)
* [Basic Example](#basic-example)
* [Debugging](#debugging)
* [Benchmark](#benchmark)
* [License](#license)

## Synopsis

```json-rules-engine``` is a powerful, lightweight rules engine.  Rules are composed of simple json structures, making them human readable and easy to persist.

## Features

* Rules expressed in simple, easy to read JSON
* Full support for ```ALL``` and ```ANY``` boolean operators, including recursive nesting
* Fast by default
* Early stopping evaluation (short-circuiting)
* Lightweight & extendable; w/few dependencies

## Installation

```bash
$ go get github.com/nimbit-software/gojson-rules-engine
```

# Contributions welcome!!!

## TODOS
- [ ] Better mascot. I am not a designer so the mascot is just a gopher with a hat
- [ ] Better error handling. Error handling can always be better. I never want it to panic but handle error gracefully
- [ ] More examples. 
- [ ] More operators. I want to add more operators to the engine
- [ ] More documentation. I want to add more documentation to the engine
- [x] Move the "rules-engine" package to the top-level for better documentation visibility.
- [X] Reduce the use of any type and create more strongly-typed methods.
- [X] Separate methods with different input types for clarity.
- [X] Optimize for speed by creating a more strongly-typed Node structure.
- [X] Implement a two-pass approach for unmarshaling and creating strongly-typed nodes.
- [ ] Add more unit tests to increase code coverage.
- [ ] Operator decorators for the rules engine.
- [ ] Create condition validation function
- [ ] Add condition sharing
- [ ] convert all rules to Json

###  ValueNode
The ValueNode is a strongly-typed node that can be used to represent any value in the rules engine. It is used to represent facts values and condition values in the rules engine. 

```go
const (
	Null DataType = iota
	Bool
	Number
	String
	Array
	Object
)

type ValueNode struct {
	Type   DataType
	Bool   bool
	Number float64
	String string
	Array  []ValueNode
	Object map[string]ValueNode
}
```

### Operators

The engine comes with the following default operators:
Either the operator itself or an alias can be used.


| Operator | Alias       | Data type           | Description                                                              | Example                                                                  |
|----------|-------------|---------------------|--------------------------------------------------------------------------|--------------------------------------------------------------------------|
| equal | eq,=        | string, number boolean |  Strict equality                                                         | ```{ "fact": "age", "operator": "equal", "value": 21 }```                |
| notEqual | ne,!=       | string, number boolean |  Strict inequality           | ```{ "fact": "age", "operator": "notEqual", "value": 21 }```             |
| in | in,contains | array               | Value is in array            | ```{ "fact": "age", "operator": "in", "value": [21, 22, 23] }```         |
| notIn | nin,notContains        | array               |  Value is not in array       | ```{ "fact": "age", "operator": "notIn", "value": [21, 22, 23] }```      |
| lessThan | lt,<        | number  | Less than                    | ```{ "fact": "age", "operator": "lessThan", "value": 21 }```             |
| lessThanInclusive | lte,<=      |  number | Less than or equal           | ```{ "fact": "age", "operator": "lessThanInclusive", "value": 21 }```    |
| greaterThan | gt,>        |  number | Greater than                 | ```{ "fact": "age", "operator": "greaterThan", "value": 21 }```          |
| greaterThanInclusive | gte,>=      |  number | Greater than or equal        | ```{ "fact": "age", "operator": "greaterThanInclusive", "value": 21 }``` |
| startsWith |             | string              | String starts with           | ```{ "fact": "name", "operator": "startsWith", "value": "B" }```         |
| endsWith |             | string              | String ends with             | ```{ "fact": "name", "operator": "endsWith", "value": "b" }```           |
| includes |             | string              | String includes              | ```{ "fact": "name", "operator": "includes", "value": "op" }```          |


Additional operators can be added via the ```AddOperator``` method.

```go
	
// func NewOperator(name string, cb func(factValue, jsonValue interface{}) bool, factValueValidator func(factValue interface{}) bool) (*Operator, error)
    o, _ := NewOperator("startsWith", func(a, b interface{}) bool {
        aString, okA := a.(string)
        bString, okB := b.(string)
        return okA && okB && strings.HasPrefix(aString, bString)
    }, nil)umberValidator), nil)

    engine.AddOperator(o, nil)

```

### Facts shared or calculated facts can be added to the engine via the ```AddFact``` or ``AddCalculatedFact`` method.

Calculated facts are facts that are calculated at runtime ONCE and then reused in the rules engine.
```go
err := engine.AddCalculatedFact("personalFoulLimit", func(a *rulesEngine.Almanac, params ...interface{}) *rulesEngine.ValueNode {
    return &rulesEngine.ValueNode{Type: rulesEngine.Number, Number: 50}
}, nil)

// or

err := engine.AddFact("test.fact", &rulesEngine.ValueNode{Type: rulesEngine.Number, Number: 50}, nil)


```


## Examples

## Basic Example

This example demonstrates an engine for detecting whether a basketball player has fouled out (a player who commits five personal fouls over the course of a 40-minute game, or six in a 48-minute game, fouls out).

```go
package main
import (
    "context"
    "encoding/json"
    "fmt"
    rulesEngine "github.com/nimbit-software/gojson-rules-engine/rulesengine"
    "os"
)

func main() {

	rule := []byte(`{
  "conditions": {
    "any": [
      {
        "all": [
          {
            "fact": "gameDuration",
            "operator": "equal",
            "value": 40
          },
          {
            "fact": "personalFoulCount",
            "operator": "greaterThanInclusive",
            "value": 5
          }
        ]
      },
      {
        "all": [
          {
            "fact": "gameDuration",
            "operator": "equal",
            "value": 48
          },
          {
            "fact": "personalFoulCount",
            "operator": "greaterThanInclusive",
            "value": 6
          }
        ]
      }
    ]
  },
  "event": {
    "type": "fouledOut",
    "params": {
      "message": "Player has fouled out!"
    }
  }
}`)

	// CONTEXT FOR EARLY-STOPPING 
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ENGINE OPTIONS
	ep := &rulesEngine.RuleEngineOptions{
		AllowUndefinedFacts: true,
	}
    // BUILD ENGINE
	engine := rulesEngine.NewEngine(nil, ep)

	// PARSE RULE
	var ruleConfig rulesEngine.RuleConfig
	if err := json.Unmarshal(rule, &ruleConfig); err != nil {
		panic(err)
	}

    // CREATE RULE
	rule, err := rulesEngine.NewRule(&ruleConfig)
	
	// ADD RULE TO ENGINE
	err = engine.AddRule(rule)

	facts := []byte(`{
            "personalFoulCount": 6,
            "gameDuration": 40,
            "name": "John",
            "user": {
                "lastName": "Jones"
            }
        }`)

    // THE ENGINE CAN RUN BOTH A MAP AND A JSON BYTE ARRAY
    res, err := engine.Run(ctx, facts)
	if err != nil {
		panic(err)
	}
    // OR 
	
	factMap := map[string]interface{}{
        "personalFoulCount": 6,
        "gameDuration": 40,
        "name": "John",
        "user": map[string]interface{}{
            "lastName": "Jones",
        },
    }
	
	res, err = engine.RunWithMap(ctx, factMap)
	if err != nil {
		panic(err)
	}
	
}	

```

More example coming soon 

## Debugging

To see what the engine is doing under the hood, debug output can be turned on via:

### Environment Variables

```bash
DEBUG=json-rules-engine
```

## Benchmarking
There is some very basic benchmarking to allow you to test the performance of the engine. 

The first test is with a single go routine and the second one with 10 go routines.

```bash
go test ./benchmarks -bench=. -run=^$ -benchmem -v 
```

Current Results 
    
```bash
# 1.000 iterations
BenchmarkRuleEngine took 35.2002ms for 1000 itterations
BenchmarkRuleEngineBasic-16                 1000             35200 ns/op            7338 B/op        108 allocs/op
BenchmarkRuleEngineWithPath took 2.9516ms for 1000 iterations
BenchmarkRuleEngineWithPath-16              1000              2952 ns/op            5595 B/op         76 allocs/op

# 10.000 iterations
BenchmarkRuleEngine took 316.1679ms for 10000 itterations
BenchmarkRuleEngineBasic-16                10000             31617 ns/op            6449 B/op        108 allocs/op
BenchmarkRuleEngineWithPath took 19.159ms for 10000 iterations
BenchmarkRuleEngineWithPath-16             10000              1916 ns/op            4930 B/op         77 allocs/op


# 100.000 iterations
BenchmarkRuleEngine took 3.2104305s for 100000 itterations
BenchmarkRuleEngineBasic-16               100000             32109 ns/op            6414 B/op        108 allocs/op
BenchmarkRuleEngineWithPath took 194.997ms for 100000 iterations
BenchmarkRuleEngineWithPath-16            100000              1950 ns/op            4847 B/op         75 allocs/op


```




## License
[ISC](./LICENSE)