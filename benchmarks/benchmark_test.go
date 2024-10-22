package benchmarks_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-faker/faker/v4"
	"os"
	"testing"
	"time"

	rulesEngine "github.com/nimbit-software/gojson-rules-engine"
)

func generateBasicTestData(n int) []map[string]interface{} {
	var testData []map[string]interface{}
	for i := 0; i < n; i++ {
		// Capture both the slice of integers and the error
		personalFoulCount, err := faker.RandomInt(0, 12)
		if err != nil {
			// Handle the error appropriately (e.g., log it or return early)
			fmt.Println("Error generating random personalFoulCount:", err)
			continue
		}
		gameDuration, err := faker.RandomInt(30, 120)
		if err != nil {
			// Handle the error appropriately (e.g., log it or return early)
			fmt.Println("Error generating random gameDuration:", err)
			continue
		}

		// Use the first integer from the slice
		testData = append(testData, map[string]interface{}{
			"personalFoulCount": personalFoulCount[0],
			"gameDuration":      gameDuration[0],
		})
	}
	return testData
}

func generateBasicTestDataByte(n int) [][]byte {
	var testData [][]byte
	for i := 0; i < n; i++ {
		// Capture both the slice of integers and the error
		personalFoulCount, err := faker.RandomInt(0, 12)
		if err != nil {
			// Handle the error appropriately (e.g., log it or return early)
			fmt.Println("Error generating random personalFoulCount:", err)
			continue
		}
		gameDuration, err := faker.RandomInt(30, 120)
		if err != nil {
			// Handle the error appropriately (e.g., log it or return early)
			fmt.Println("Error generating random gameDuration:", err)
			continue
		}

		data := map[string]interface{}{
			"personalFoulCount": personalFoulCount[0],
			"gameDuration":      gameDuration[0],
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshaling test data to JSON:", err)
			continue
		}

		testData = append(testData, jsonData)
	}
	return testData
}

func BenchmarkRuleEngineBasic(b *testing.B) {
	jsonBytes, err := os.ReadFile("../examples/game_foul_rule.json")
	if err != nil {
		b.Fatalf("Failed to read rule file: %v", err)
	}

	var ruleConfig rulesEngine.RuleConfig
	if err := json.Unmarshal(jsonBytes, &ruleConfig); err != nil {
		b.Fatalf("Failed to unmarshal rule JSON: %v", err)
	}

	testData := generateBasicTestData(b.N)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	engine := rulesEngine.NewEngine(nil, &rulesEngine.RuleEngineOptions{
		AllowUndefinedFacts: true,
	})
	rule, err := rulesEngine.NewRule(&ruleConfig)
	err = engine.AddRule(rule)

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := engine.RunWithMap(ctx, testData[i%len(testData)])
		if err != nil {
			b.Fatalf("Engine run failed: %v", err)
		}
	}
	elapsed := time.Since(start)
	b.Logf("BenchmarkRuleEngine took %s for %v itterations", elapsed, b.N)
}
