package src

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
