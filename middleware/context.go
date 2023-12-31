package middleware

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var ErrInvalidContext = errors.New("invalid context")

type echoContext = echo.Context

type Context interface {
	echoContext

	GoContext() context.Context
	SetContext(ctx context.Context)
}

type customContext struct {
	echoContext

	ctx context.Context
}

func NewContext(echoCtx echo.Context) Context {
	return &customContext{
		echoContext: echoCtx,
	}
}

func (c *customContext) SetContext(ctx context.Context) {
	if ctx == nil {
		return
	}

	c.ctx = ctx
}

func (c *customContext) GoContext() context.Context {
	if c.ctx == nil {
		c.ctx = c.Request().Context()
	}

	return c.ctx
}

func UpgradeContext(echoCtx echo.Context) (Context, error) {
	ctx, ok := echoCtx.(Context)
	if !ok {
		return nil, ErrInvalidContext
	}
	return ctx, nil
}

func CustomContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		logrus.Debug("Upgrade context")
		customCtx := NewContext(echoCtx)
		return next(customCtx)
	}
}