package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bigelle/ratebucket"
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
		c.Response().After(func() {
			end := time.Since(start)

			code := c.Response().Status
			if code >= 400 {
				app.Logger.Error("error answering request",
					zap.Int("code", c.Response().Status),
					zap.String("message", http.StatusText(code)),
					zap.String("path", c.Request().URL.Path),
					zap.Duration("duration", end),
				)
				return
			}

			app.Logger.Info(http.StatusText(c.Response().Status),
				zap.Int("code", c.Response().Status),
				zap.String("path", c.Request().URL.Path),
				zap.Duration("duration", end),
			)
		})

		return next(c)
	}
}

type RateLimiter struct {
	*ratebucket.Pool
}

func (r *RateLimiter) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !r.Allow(clientIP(c.Request())) {
			return echo.ErrTooManyRequests
		}
		return next(c)
	}
}

func clientIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	if xr := req.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}
