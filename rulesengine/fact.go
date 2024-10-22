package rulesengine

import (
	"github.com/tidwall/gjson"
	"sync"
)

// FactMap is a thread-safe map used to store and manage facts in the rules engine.
// It provides methods for setting, loading, and deleting facts, as well as iterating over the map.
type FactMap struct {
	internalMap sync.Map
}

// Set stores a fact in the FactMap using the hashed key for efficient lookup.
// Params:
// - key: The key to associate with the fact.
// - value: The fact to store.
func (m *FactMap) Set(key string, value *Fact) {
	hash := HashString(key)
	m.internalMap.Store(hash, value)
}

// Load retrieves a fact from the FactMap using the hashed key.
// Params:
// - key: The key associated with the fact.
// Returns:
// - A pointer to the Fact, and a boolean indicating whether the fact was found.
func (m *FactMap) Load(key string) (*Fact, bool) {
	hash := HashString(key)
	value, ok := m.internalMap.Load(hash)
	if !ok {
		return &Fact{}, false
	}
	return value.(*Fact), ok
}

// LoadOrStore retrieves a fact from the FactMap if it exists, or stores the provided fact if it does not.
// Params:
// - key: The key to associate with the fact.
// - value: The fact to store if the key does not exist.
// Returns:
// - A pointer to the actual fact (either loaded or newly stored), and a boolean indicating if it was already present.
func (m *FactMap) LoadOrStore(key string, value *Fact) (*Fact, bool) {
	hash := HashString(key)
	actualValue, loaded := m.internalMap.LoadOrStore(hash, value)
	return actualValue.(*Fact), loaded
}

// Delete removes a fact from the FactMap using the hashed key.
// Params:
// - key: The key associated with the fact to be removed.
func (m *FactMap) Delete(key string) {
	hash := HashString(key)
	m.internalMap.Delete(hash)
}

// Range iterates over all key-value pairs in the FactMap, applying the provided function to each.
// The function should return true to continue iteration, or false to stop.
// Params:
// - f: The function to apply to each key-value pair.
func (m *FactMap) Range(f func(key string, value *Fact) bool) {
	rawFunc := func(key uint64, value *Fact) bool {
		return f(value.Path, value)
	}
	m.internalMap.Range(func(path, value interface{}) bool {
		return rawFunc(path.(uint64), value.(*Fact))
	})
}

// NewValueFromGjson converts a gjson.Result into a ValueNode.
// It handles various data types such as null, string, number, boolean, and arrays.
// Params:
// - result: The gjson.Result to be converted.
// Returns a pointer to a ValueNode representing the result.
func NewValueFromGjson(result gjson.Result) *ValueNode {
	switch result.Type {
	case gjson.Null:
		return &ValueNode{Type: Null}
	case gjson.String:
		return &ValueNode{Type: String, String: result.String()}
	case gjson.Number:
		return &ValueNode{Type: Number, Number: result.Float()}
	case gjson.True, gjson.False:
		return &ValueNode{Type: Bool, Bool: result.Bool()}
	case gjson.JSON:
		if result.IsArray() {
			arrayValues := make([]ValueNode, 0)
			result.ForEach(func(_, value gjson.Result) bool {
				v := NewValueFromGjson(value)
				arrayValues = append(arrayValues, *v)
				return true // Continue iteration
			})
			return &ValueNode{Type: Array, Array: arrayValues}
		} else {
			// Handle objects if needed
			return &ValueNode{Type: Null}
		}
	default:
		return &ValueNode{Type: Null}
	}
}

// Fact represents a fact within the rules engine.
// It holds a value (as a ValueNode), a path identifying the fact, and optional metadata about how the value was calculated.
type Fact struct {
	Value             *ValueNode
	Path              string
	CalculationMethod DynamicFactCallback
	Cached            bool
	Priority          int
	Dynamic           bool
}

// NewCalculatedFact creates a new Fact instance with a dynamic calculation method.
// Params:
// path: The path identifying the fact.
// method: The method to calculate the fact value.
// options: Optional configuration options for the fact.
func NewCalculatedFact(path string, method DynamicFactCallback, options *FactOptions) *Fact {
	defaultOptions := FactOptions{Cache: true, Priority: 1}
	if options == nil {
		options = &defaultOptions
	}

	return &Fact{
		Priority:          options.Priority,
		Cached:            options.Cache,
		Path:              path,
		CalculationMethod: method,
		Dynamic:           true,
	}
}

// NewFact creates a new Fact instance with a static value.
// Params:
// path: The path identifying the fact.
// value: The value of the fact.
// options: Optional configuration options for the fact.
func NewFact(path string, value ValueNode, options *FactOptions) (*Fact, error) {
	defaultOptions := FactOptions{Cache: true, Priority: 1}
	if options == nil {
		options = &defaultOptions
	}

	return &Fact{
		Value:    &value,
		Priority: options.Priority,
		Cached:   options.Cache,
		Dynamic:  false,
		Path:     path,
	}, nil
}

// Calculate evaluates the fact value using the provided Almanac and optional parameters.
// If the fact is dynamic, it uses the calculation method to determine the value.
// Params:
// almanac: The Almanac instance to use for calculation.
// params: Optional parameters to pass to the calculation method.
func (f *Fact) Calculate(almanac *Almanac, params ...interface{}) *Fact {
	if f.Dynamic {
		f.Value = f.CalculationMethod(almanac, params...)
		return f
	}
	// TODO USE ALMANAC TO CALCULATE FACT VALUE
	return f
}
