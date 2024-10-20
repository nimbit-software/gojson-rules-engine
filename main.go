package main

import (
	"context"
	"encoding/json"
	"fmt"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/rulesengine"
)

func main() {

	ruleRaw := []byte(`{
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

	var ruleConfig rulesEngine.RuleConfig
	if err := json.Unmarshal(ruleRaw, &ruleConfig); err != nil {
		panic(err)
	}

	engine := rulesEngine.NewEngine(nil, ep)

	rule, err := rulesEngine.NewRule(&ruleConfig)
	err = engine.AddRule(rule)

	facts := []byte(`{
            "personalFoulCount": 7,
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
