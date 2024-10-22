package rulesengine

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type DataType int

const (
	Null DataType = iota
	Bool
	Number
	String
	Array
	Object
)

// ValueNode represents a value used in conditions and comparisons.
// It supports types such as strings, numbers, booleans, arrays, and null.
type ValueNode struct {
	Type   DataType
	Bool   bool
	Number float64
	String string
	Array  []ValueNode
	Object map[string]ValueNode
}

func (v *ValueNode) IsArray() bool {
	return v.Type == Array
}

func (v *ValueNode) IsObject() bool {
	return v.Type == Object
}

func (v *ValueNode) IsNull() bool {
	return v.Type == Null
}

func (v *ValueNode) IsBool() bool {
	return v.Type == Bool
}

func (v *ValueNode) IsNumber() bool {
	return v.Type == Number
}

func (v *ValueNode) IsString() bool {
	return v.Type == String
}

func (v *ValueNode) SameType(other *ValueNode) bool {
	return v.Type == other.Type
}

func (v *ValueNode) Raw() interface{} {
	switch v.Type {
	case Null:
		return nil
	case Bool:
		return v.Bool
	case Number:
		return v.Number
	case String:
		return v.String
	case Array:
		rawArray := make([]interface{}, len(v.Array))
		for i, item := range v.Array {
			rawArray[i] = item.Raw()
		}
		return rawArray
	case Object:
		rawObject := make(map[string]interface{})
		for key, value := range v.Object {
			rawObject[key] = value.Raw()
		}
		return rawObject
	default:
		return nil
	}
}

func (v *ValueNode) UnmarshalJSON(data []byte) error {
	// Remove leading and trailing whitespace
	data = bytes.TrimSpace(data)

	// Handle null
	if bytes.Equal(data, []byte("null")) {
		v.Type = Null
		return nil
	}

	// Handle boolean
	if bytes.Equal(data, []byte("true")) {
		v.Type = Bool
		v.Bool = true
		return nil
	}
	if bytes.Equal(data, []byte("false")) {
		v.Type = Bool
		v.Bool = false
		return nil
	}

	// Handle number
	if len(data) > 0 && (data[0] == '-' || (data[0] >= '0' && data[0] <= '9')) {
		var num float64
		if err := json.Unmarshal(data, &num); err == nil {
			v.Type = Number
			v.Number = num
			return nil
		}
	}

	// Handle string
	if len(data) > 0 && data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err == nil {
			v.Type = String
			v.String = str
			return nil
		}
	}

	// Handle array
	if len(data) > 0 && data[0] == '[' {
		v.Type = Array
		var rawArray []json.RawMessage
		if err := json.Unmarshal(data, &rawArray); err != nil {
			return err
		}
		v.Array = make([]ValueNode, len(rawArray))
		for i, item := range rawArray {
			if err := v.Array[i].UnmarshalJSON(item); err != nil {
				return fmt.Errorf("error unmarshaling array element %d: %v", i, err)
			}
		}
		return nil
	}

	// Handle object
	if len(data) > 0 && data[0] == '{' {
		v.Type = Object
		var rawObject map[string]json.RawMessage
		if err := json.Unmarshal(data, &rawObject); err != nil {
			return err
		}
		v.Object = make(map[string]ValueNode)
		for key, rawValue := range rawObject {
			var child ValueNode
			if err := child.UnmarshalJSON(rawValue); err != nil {
				return fmt.Errorf("error unmarshaling object field '%s': %v", key, err)
			}
			v.Object[key] = child
		}
		return nil
	}

	// If none of the above, return an error
	return fmt.Errorf("unknown JSON value: %s", string(data))
}
