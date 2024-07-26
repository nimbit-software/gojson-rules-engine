package rulesengine

import (
	"errors"
	"github.com/mitchellh/hashstructure/v2"
)

const (
	CONSTANT = "CONSTANT"
	DYNAMIC  = "DYNAMIC"
)

type Fact struct {
	ID                string
	Value             interface{}
	CalculationMethod DynamicFactCallback
	Type              string
	Priority          int
	Options           FactOptions
	CacheKeyMethod    func(id string, params map[string]interface{}) map[string]interface{}
}

func NewFact(id string, valueOrMethod interface{}, options *FactOptions) (*Fact, error) {
	if id == "" {
		return nil, errors.New("factId required")
	}

	defaultOptions := FactOptions{Cache: true, Priority: 1}
	if options == nil {
		options = &defaultOptions
	}

	fact := &Fact{
		ID:             id,
		Priority:       options.Priority,
		Options:        *options,
		CacheKeyMethod: defaultCacheKeys,
	}

	if method, ok := valueOrMethod.(func(params map[string]interface{}, almanac *Almanac) interface{}); ok {
		fact.CalculationMethod = method
		fact.Type = DYNAMIC
	} else {
		fact.Value = valueOrMethod
		fact.Type = CONSTANT
	}

	return fact, nil
}

func (f *Fact) IsConstant() bool {
	return f.Type == CONSTANT
}

func (f *Fact) IsDynamic() bool {
	return f.Type == DYNAMIC
}

func (f *Fact) Calculate(params map[string]interface{}, almanac *Almanac) interface{} {
	if f.Type == CONSTANT {
		return f.Value
	}
	return f.CalculationMethod(params, almanac)
}

func HashFromObject(obj interface{}) (uint64, error) {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, nil)
	if err != nil {
		return 0, err
	}
	return hash, nil
}

func defaultCacheKeys(id string, params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"params": params,
		"id":     id,
	}
}

func (f *Fact) GetCacheKey(params map[string]interface{}) (uint64, error) {
	if f.Options.Cache {
		cacheProperties := f.CacheKeyMethod(f.ID, params)
		hash, err := HashFromObject(cacheProperties)
		if err != nil {
			return 0, err
		}
		return hash, nil
	}
	return 0, nil
}
