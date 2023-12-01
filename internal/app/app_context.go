package app

import "context"

type Context interface {
	context.Context
	UserId() string
	RequestId() string
	CausationId() string
	CorrelationId() string
}

type appContext struct {
	context.Context
	userId        string
	requestId     string
	causationId   string
	correlationId string
}

func (c *appContext) UserId() string {
	return c.userId
}

func (c *appContext) RequestId() string {
	return c.requestId
}

func (c *appContext) CausationId() string {
	return c.causationId
}

func (c *appContext) CorrelationId() string {
	return c.correlationId
}

func NewContext(ctx context.Context, userId string, requestId string, causationId string, correlationId string) Context {
	return &appContext{
		Context:       ctx,
		userId:        userId,
		requestId:     requestId,
		causationId:   causationId,
		correlationId: correlationId,
	}
}
