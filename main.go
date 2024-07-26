package main

import (
	"context"
	"encoding/json"
	"fmt"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/rulesengine"
	"os"
)

func main() {
	//jsonBytes, err := os.ReadFile("examples/game_foul_rule.json")
	jsonBytes, err := os.ReadFile("examples/endsWith-rule.json")
	if err != nil {
		panic(err)
	}

	var ruleMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &ruleMap); err != nil {
		panic(err)
	}

	data := map[string]interface{}{
		"personalFoulCount": 1,
		"gameDuration":      40,
		"user": map[string]interface{}{
			"lastName":  "Sooter",
			"firstName": "David",
		},
	}

	dataByte := []byte(`{
        "personalFoulCount": 6,
        "gameDuration": 40,
		"name": "John",
        "user": {
			"firstName": "David",
            "lastName": "Sooter"
        }
    }`)

	ruleMap["onSuccess"] = func(ruleResult *rulesEngine.RuleResult) {
		fmt.Println("Rule succeeded")
	}

	ruleMap["onFailure"] = func(ruleResult *rulesEngine.RuleResult) {
		fmt.Println("Rule succeeded")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine := rulesEngine.NewEngine(nil, &rulesEngine.RuleEngineOptions{
		AllowUndefinedFacts:       true,
		ReplaceFactsInEventParams: true,
	})

	engine.AddFact("fullName", func(params map[string]interface{}, almanac *rulesEngine.Almanac) interface{} {
		almanac.GetValue("user.firstName")
		return "bla"
	}, nil)

	engine.AddRule(ruleMap)
	json, _ := engine.GetRules()[0].ToJSON(false)
	rulesEngine.NewRule(json)
	fmt.Printf("%+v\n", json)

	res, err := engine.Run(ctx, dataByte)
	printResults(res)
	res, err = engine.Run(ctx, data)

	printResults(res)
}

func printResults(res map[string]interface{}) {
	fmt.Println("Results:")
	for _, result := range res["results"].([]*rulesEngine.RuleResult) {
		fmt.Printf("%+v\n", result)
	}

	fmt.Println("Failure Results:")
	for _, failureResult := range res["failureResults"].([]*rulesEngine.RuleResult) {
		fmt.Printf("%+v\n", failureResult)
	}

	fmt.Println("Almanac:")
	fmt.Printf("%+v\n", res["almanac"])

	fmt.Println("Events:")
	fmt.Printf("%+v\n", res["events"])

	fmt.Println("Failure Events:")
	fmt.Printf("%+v\n", res["failureEvents"])
}
