package benchmarks_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	faker "github.com/go-faker/faker/v4"
	rulesEngine "github.com/nimbit-software/gojson-rules-engine/rulesengine"
)

func generateTestData(n int) []map[string]interface{} {
	var testData []map[string]interface{}
	for i := 0; i < n; i++ {
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

func generateTestDataByte(n int) [][]byte {
	var testData [][]byte
	for i := 0; i < 1000000; i++ {
		data := map[string]interface{}{
			"personalFoulCount": i % 10,
			"gameDuration":      i % 48,
			"user": map[string]interface{}{
				"lastName": faker.LastName(),
			},
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

func BenchmarkRuleEngineWithPath(b *testing.B) {
	jsonBytes, err := os.ReadFile("../examples/endsWith-rule.json")
	if err != nil {
		b.Fatalf("Failed to read rule file: %v", err)
	}

	var ruleConfig rulesEngine.RuleConfig
	if err := json.Unmarshal(jsonBytes, &ruleConfig); err != nil {
		b.Fatalf("Failed to unmarshal rule JSON: %v", err)
	}

	testDataByte := generateTestDataByte(b.N)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engine := rulesEngine.NewEngine(nil, &rulesEngine.RuleEngineOptions{
		AllowUndefinedFacts: true,
	})

	rule, err := rulesEngine.NewRule(&ruleConfig)
	err = engine.AddRule(rule)

	b.ResetTimer()
	start := time.Now()

	// Use goroutines to parallelize the benchmarking process
	numGoroutines := 10
	var wg sync.WaitGroup
	chunkSize := b.N / numGoroutines

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			startIndex := g * chunkSize
			endIndex := startIndex + chunkSize
			if g == numGoroutines-1 { // Handle the remainder
				endIndex = b.N
			}
			for i := startIndex; i < endIndex; i++ {
				_, err := engine.Run(ctx, testDataByte[i])
				if err != nil {
					b.Fatalf("Engine run failed: %v", err)
				}
			}
		}(g)
	}

	wg.Wait() // Wait for all goroutines to finish

	elapsed := time.Since(start)
	b.Logf("BenchmarkRuleEngineWithPath took %s for %v iterations", elapsed, b.N)
}
