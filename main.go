package main

import (
	"context"
	"encoding/json"
	"fmt"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/src"
	"os"
	"time"
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
		"personalFoulCount": 1,
		"gameDuration":      40,
		"user": map[string]interface{}{
			"lastName": "Sooter",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	engine := rulesEngine.NewEngine(nil, nil)

	engine.AddRule(ruleMap)

	res, err := engine.Run(ctx, cart)

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
