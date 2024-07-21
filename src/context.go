package src

import (
	"context"
)

type EvaluationContext struct {
	context.Context
	Errors []error // TODO CUSTOM ERROR TYPE
}

func NewEvaluationContext(ctx context.Context) *EvaluationContext {
	return &EvaluationContext{
		Context: ctx,
		Errors:  []error{},
	}
}

func (c *EvaluationContext) AddError(err error) {
	c.Errors = append(c.Errors, err)
}
