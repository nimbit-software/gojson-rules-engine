package main

import (
	"context"
	"encoding/json"
	"fmt"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/rulesengine"
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
