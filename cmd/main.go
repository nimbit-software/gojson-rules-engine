package main

import (
	"context"
	"encoding/json"
	"fmt"
	re "github.com/nimbit-software/gojson-rules-engine"
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
          }
        ],
		"all": [
			{
			"priority": 10,
            "fact": "personalFoulLimit",
            "operator": ">",
            "value": 60
          },
		 {
			"priority": 10,
            "fact": "personalFoulLimit",
            "operator": "<",
            "value": 60
          }
		]
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
	ep := &re.RuleEngineOptions{
		AllowUndefinedFacts:       true,
		ReplaceFactsInEventParams: true,
		AllowUndefinedConditions:  true,
	}

	var ruleConfig re.RuleConfig
	if err := json.Unmarshal(ruleRaw, &ruleConfig); err != nil {
		panic(err)
	}

	engine := re.NewEngine(nil, ep)

	err := engine.AddCalculatedFact("personalFoulLimit", func(a *re.Almanac, params ...interface{}) *re.ValueNode {
		return &re.ValueNode{Type: re.Number, Number: 50}
	}, nil)

	err = engine.AddFact("test.fact", &re.ValueNode{Type: re.Number, Number: 50}, nil)

	rule, err := re.NewRule(&ruleConfig)
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

	res, err := engine.Run(ctx, facts)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
