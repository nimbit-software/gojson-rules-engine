package main

import (
	"encoding/json"
	"fmt"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/src"
	"os"
)

func main() {
	jsonBytes, err := os.ReadFile("examples/endsWith-rule.json")
	if err != nil {
		panic(err)
	}

	var ruleMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &ruleMap); err != nil {
		panic(err)
	}

	cart := map[string]interface{}{
		"personalFoulCount": 6,
		"gameDuration":      40,
		"user": map[string]interface{}{
			"lastName": "Sooter",
		},
	}

	engine := rulesEngine.NewEngine(nil, nil)

	engine.AddRule(ruleMap)

	res, err := engine.Run(cart)

	fmt.Printf("%+v", res)
}
