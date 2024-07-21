package src

import (
	"fmt"
	"os"
	"strings"
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

// isDebugMode checks if the DEBUG environment variable contains "json-rules-engine"
func isDebugMode() bool {
	debugEnv, debugEnvExists := os.LookupEnv("DEBUG")
	if debugEnvExists && strings.Contains(debugEnv, "json-rules-engine") {
		return true
	}

	// Optionally, handle local storage or other debugging mechanisms here

	return false
}
