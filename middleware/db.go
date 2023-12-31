package middleware

import (
	"buddle-server/internal/db"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func WithDB(key interface{}, conn *gorm.DB, opts ...DatabaseOption) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			switch {
			case key == nil:
				return fmt.Errorf("nil db key")
			case conn == nil:
				return fmt.Errorf("nil db conn")
			}

			for _, opt := range opts {
				conn = opt(conn)
			}

			ctx, err := UpgradeContext(c)
			if err != nil {
				return errors.Wrap(err, "upgrade context")
			}
			ctx.SetContext(db.ContextWithConn(ctx.GoContext(), db.WriteDBKey, conn))

			return next(ctx)
		}
	}
}

type DatabaseOption func(conn *gorm.DB) *gorm.DB