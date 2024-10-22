package rulesengine

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	debugMode     bool
	debugModeOnce sync.Once
)

// Debug logs the message if the DEBUG environment variable contains "json-rules-engine"
func Debug(message string) {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic, do nothing
		}
	}()

	if isDebugMode() {
		fmt.Println(message)
	}
}

func isDebugMode() bool {
	debugModeOnce.Do(func() {
		debugEnv, debugEnvExists := os.LookupEnv("DEBUG")
		if debugEnvExists && strings.Contains(debugEnv, "json-rules-engine") {
			debugMode = true
		}
	})
	return debugMode
}
