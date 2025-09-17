package handlers

import (
	"net/http"
	"time"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Database struct {
	Queries *database.Queries
	Conn    *pgx.Conn
}

type Config struct {
	JWTAccessSecret  []byte
	JWTRefreshSecret []byte
}

type App struct {
	DB     Database
	Logger *zap.Logger
	Config Config
}

func (app App) LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		end := time.Since(start)
		if err != nil {
			app.Logger.Error("error answering request",
				zap.Int("code", c.Response().Status),
				zap.String("path", c.Path()),
				zap.Duration("duration", end),
				zap.Error(err),
			)
			return nil
		}

		app.Logger.Info(http.StatusText(c.Response().Status),
			zap.Int("code", c.Response().Status),
			zap.String("path", c.Path()),
			zap.Duration("duration", end),
		)
		return nil
	}
}
