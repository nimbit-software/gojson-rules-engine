package benchmarks_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	faker "github.com/bxcodec/faker/v3"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/src"
)

func generateTestData() []map[string]interface{} {
	var testData []map[string]interface{}
	for i := 0; i < 1000000; i++ {
		testData = append(testData, map[string]interface{}{
			"personalFoulCount": i % 10,
			"gameDuration":      i % 48,
			"user": map[string]interface{}{
				"lastName": faker.LastName(),
			},
		})
	}
	return testData
}

func BenchmarkRuleEngineWithPath(b *testing.B) {
	jsonBytes, err := os.ReadFile("../examples/endsWith-rule.json")
	if err != nil {
		b.Fatalf("Failed to read rule file: %v", err)
	}

	var ruleMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &ruleMap); err != nil {
		b.Fatalf("Failed to unmarshal rule JSON: %v", err)
	}

	testData := generateTestData()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engine := rulesEngine.NewEngine(nil, &rulesEngine.RuleEngineOptions{
		AllowUndefinedFacts: true,
	})
	engine.AddRule(ruleMap)

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := engine.Run(ctx, testData[i%len(testData)])
		if err != nil {
			b.Fatalf("Engine run failed: %v", err)
		}
	}
	elapsed := time.Since(start)
	b.Logf("BenchmarkRuleEngineWithPath took %s for %v itterations", elapsed, b.N)
}
