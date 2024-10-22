package rulesengine

import (
	"context"
)

// ExecutionContext holds metadata and control flags for rule execution.
type ExecutionContext struct {
	context.Context
	Cancel    context.CancelFunc
	StopEarly bool
	Message   string
	Errors    []error
}

func NewEvaluationContext(ctx context.Context) *ExecutionContext {
	return &ExecutionContext{
		Context: ctx,
		Errors:  []error{},
	}
}

func (c *ExecutionContext) AddError(err error) {
	c.Errors = append(c.Errors, err)
}
