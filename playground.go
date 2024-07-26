package main

//
//import (
//	"fmt"
//	"github.com/buger/jsonparser"
//	"log"
//	"sync"
//	"time"
//)
//
//func main() {
//	// Generate a million nested JSON objects
//	// Create a single JSON object
//	jsonData := []byte(`{"name": "John", "age": 30, "address": {"city": "New York"}}`)
//
//	// Number of objects to process
//	numObjects := 100000000
//
//	// Benchmark Direct Access using jsonparser
//	start := time.Now()
//	for i := 0; i < numObjects; i++ {
//		_, err := jsonparser.GetString(jsonData, "address", "city")
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//	jsonParserDuration := time.Since(start)
//	fmt.Printf("Direct access using jsonparser duration: %v\n", jsonParserDuration)
//
//	// Benchmark Concurrent Direct Access using jsonparser
//	start = time.Now()
//	var wg sync.WaitGroup
//	numWorkers := 4 // Adjust based on the number of CPU cores
//	chunkSize := numObjects / numWorkers
//
//	for i := 0; i < numWorkers; i++ {
//		wg.Add(1)
//		go func(startIdx int) {
//			defer wg.Done()
//			for j := startIdx; j < startIdx+chunkSize; j++ {
//				_, err := jsonparser.GetString(jsonData, "address", "city")
//				if err != nil {
//					log.Fatal(err)
//				}
//			}
//		}(i * chunkSize)
//	}
//	wg.Wait()
//	concurrentJsonParserDuration := time.Since(start)
//	fmt.Printf("Concurrent direct access using jsonparser duration: %v\n", concurrentJsonParserDuration)
//}
