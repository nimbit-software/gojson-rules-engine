package rulesengine

import (
	"hash/fnv"
	"reflect"
)

// isObjectLike checks if the value is an object-like structure
func IsObjectLike(value interface{}) bool {
	return value != nil && reflect.ValueOf(value).Kind() == reflect.Map
}

func HashString(data string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(data))
	return h.Sum64()
}
