package rulesengine

import (
	"encoding/json"
	"reflect"
)

// isObjectLike checks if the value is an object-like structure
func IsObjectLike(value interface{}) bool {
	return value != nil && reflect.ValueOf(value).Kind() == reflect.Map
}

func DeepCopy(src, dst interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dst)
}

func ParsePriority(properties map[string]interface{}) (int, *InvalidRuleError) {
	var result int
	var err *InvalidRuleError

	if _, exists := properties["priority"]; exists {
		switch priority := properties["priority"].(type) {
		case float64:
			result = int(priority)
		case float32:
			result = int(priority)
		case int64:
			result = int(priority)
		case int:
			result = priority
		default:
			err = NewInvalidPriorityTypeError()
		}

		if result <= 0 {
			err = NewInvalidPriorityValueError()
		}
	} else {
		err = NewPriorityNotSetError()
	}

	return result, err
}
