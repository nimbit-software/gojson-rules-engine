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
        "any": [
          {
			"priority": 10,
            "fact": "gameDuration",
            "operator": "equal",
            "value": 40
          },
          {
			"priority": 11,
            "fact": "use.lastName",
            "operator": "=",
            "value": "Jones"
          }
        ]
      }
    ],
	"all": [
		{
			"fact": "user.lastName",
			"operator": "includes",
			"value": "Jo"
		}
	]
  },
  "event": {
    "type": "fouledOut",
    "params": {
      "firstName": {
        "fact": "user.lastName"
      },
      "message": "Player has fouled out!"
    }
  }
}`)

	// CONTEXT FOR EARLY-STOPPING
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ENGINE OPTIONS
	ep := &rulesEngine.RuleEngineOptions{
		AllowUndefinedFacts:       true,
		ReplaceFactsInEventParams: true,
		AllowUndefinedConditions:  true,
	}

	var ruleConfig rulesEngine.RuleConfig
	if err := json.Unmarshal(ruleRaw, &ruleConfig); err != nil {
		panic(err)
	}

	engine := rulesEngine.NewEngine(nil, ep)

	rule, err := rulesEngine.NewRule(&ruleConfig)
	err = engine.AddRule(rule)

	facts := []byte(`{
            "personalFoulCount": 4,
            "gameDuration": 40,
            "name": "John",
            "user": {
                "lastName": "Jones",
				"gameDuration": 40
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
