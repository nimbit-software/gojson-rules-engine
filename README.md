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
- [ ] Better Types. I copied and transformed the code but a lot of it works with interface{} which is not ideal
- [ ] Better error handling. Error handling can always be better. I never want it to panic but handle error gracefully
- [ ] Testing. Almost no test atm just a few POC benchmarks
- [ ] More examples. I want to add more examples to the README
- [ ] More operators. I want to add more operators to the engine
- [ ] More documentation. I want to add more documentation to the engine
- [ ] More performance. It is fast (faster than the node version and even faster if you give it some more threads). But i think there is room for improvement especially on how the json data is accessed. 
- [ ] There are some of the more advance features that are ported but not all tested yet.



## Operators

The engine comes with the following default operators:
Either the operator itself or an alias can be used.


| Operator | Alias  | Data type           | Description                                                              | Example                                                                  |
|----------|--------|---------------------|--------------------------------------------------------------------------|--------------------------------------------------------------------------|
| equal | eq,=   | string, number boolean |  Strict equality                                                         | ```{ "fact": "age", "operator": "equal", "value": 21 }```                |
| notEqual | ne,!=  | string, number boolean |  Strict inequality           | ```{ "fact": "age", "operator": "notEqual", "value": 21 }```             |
| in | in     | array               | Value is in array            | ```{ "fact": "age", "operator": "in", "value": [21, 22, 23] }```         |
| notIn | nin    | array               |  Value is not in array       | ```{ "fact": "age", "operator": "notIn", "value": [21, 22, 23] }```      |
| lessThan | lt,<   | number  | Less than                    | ```{ "fact": "age", "operator": "lessThan", "value": 21 }```             |
| lessThanInclusive | lte,<= |  number | Less than or equal           | ```{ "fact": "age", "operator": "lessThanInclusive", "value": 21 }```    |
| greaterThan | gt,>   |  number | Greater than                 | ```{ "fact": "age", "operator": "greaterThan", "value": 21 }```          |
| greaterThanInclusive | gte,>= |  number | Greater than or equal        | ```{ "fact": "age", "operator": "greaterThanInclusive", "value": 21 }``` |
| contains |  | array               | Array contains value         | ```{ "fact": "names", "operator": "contains", "value": "Bob" }```        |
| notContains |  | array               |  Array does not contain value | ```{ "fact": "names", "operator": "notContains", "value": "Bob" }```     |
| startsWith |  | string              | String starts with           | ```{ "fact": "name", "operator": "startsWith", "value": "B" }```         |
| endsWith |  | string              | String ends with             | ```{ "fact": "name", "operator": "endsWith", "value": "b" }```           |
| includes |  | string              | String includes              | ```{ "fact": "name", "operator": "includes", "value": "op" }```          |


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

	var ruleMap map[string]interface{}
	if err := json.Unmarshal(rule, &ruleMap); err != nil {
		panic(err)
	}

	engine := rulesEngine.NewEngine(nil, ep)

	engine.AddRule(ruleMap)

	facts := []byte(`{
            "personalFoulCount": 6,
            "gameDuration": 40,
            "name": "John",
            "user": {
                "lastName": "Jones"
            }
        }`)

	// THE RUN FUNCTION ACCEPTS BOTH A MAP AND A BYTE ARRAY 
	// - []byte (byte array offers slightly better performance) under the hood github.com/buger/jsonparser is used to parse it into the almanac
	// - map[string]interface{}
	res, err := engine.Run(ctx, facts)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
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

```bash
go test ./benchmarks -bench=. -run=^$ -benchmem -v 
```

## License
[ISC](./LICENSE)