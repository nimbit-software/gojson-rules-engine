package src

type Event struct {
	Type   string
	Params map[string]interface{}
}

type FactOptions struct {
	Cache    bool
	Priority int
}

type DynamicFactCallback func(params map[string]interface{}, almanac *Almanac) interface{}

type EvaluationResult struct {
	Result             bool        `json:"Result"`
	LeftHandSideValue  interface{} `json:"LeftHandSideValue"`
	RightHandSideValue interface{} `json:"RightHandSideValue"`
	Operator           string      `json:"Operator"`
}
