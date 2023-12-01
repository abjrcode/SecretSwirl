package app

import (
	"context"

	"github.com/rs/zerolog"
)

type Context interface {
	context.Context

	Logger() *zerolog.Logger

	UserId() string
	RequestId() string
	CausationId() string
	CorrelationId() string
}

type appContext struct {
	context.Context
	logger        *zerolog.Logger
	userId        string
	requestId     string
	causationId   string
	correlationId string
}

func (appCtx *appContext) Logger() *zerolog.Logger {
	return appCtx.logger
}

func (appCtx *appContext) UserId() string {
	return appCtx.userId
}

func (appCtx *appContext) RequestId() string {
	return appCtx.requestId
}

func (appCtx *appContext) CausationId() string {
	return appCtx.causationId
}

func (appCtx *appContext) CorrelationId() string {
	return appCtx.correlationId
}

func NewContext(ctx context.Context, userId string, requestId string, causationId string, correlationId string, logger *zerolog.Logger) Context {
	return &appContext{
		Context:       ctx,
		logger:        logger,
		userId:        userId,
		requestId:     requestId,
		causationId:   causationId,
		correlationId: correlationId,
	}
}
