package main

import (
	"context"
	"os"

	"github.com/bigelle/ratebucket"
	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/handlers"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	// LOGGER:
	logger, _ := zap.NewProduction(
		zap.AddStacktrace(zap.FatalLevel),
		zap.WithCaller(false),
	)
	defer logger.Sync()

	// ENVIRONMENT:
	if err := godotenv.Load(".env"); err != nil {
		logger.Fatal("failed to load environment variables", zap.Error(err))
	}

	// DATABASE:
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("failed to connect to the database", zap.Error(err))
	}
	defer conn.Close(ctx)
	queries := database.New(conn)

	// RATE LIMITER:
	authRL := handlers.RateLimiter{
		Pool: ratebucket.NewPoolConfig(ratebucket.PoolConfig{
			InitialTokens: 5,
			Capacity:      5,
			RefillRate:    1.0 / 12, // 1 = 60/min, 1\12 = 5/min
		}),
	}

	RL := handlers.RateLimiter{
		Pool: ratebucket.NewPoolConfig(ratebucket.PoolConfig{
			InitialTokens: 100,
			Capacity:      100,
			RefillRate:    1.0,
		}),
	}

	// APP:
	app := handlers.App{
		DB: handlers.Database{
			Conn:    conn,
			Queries: queries,
		},
		Logger: logger,
	}

	// ROUTER:
	r := echo.New()
	r.Use(
		middleware.Recover(),
		app.LoggingMiddleware,
	)
	r.Pre(middleware.RemoveTrailingSlash())

	// Unprotected routes:
	auth := r.Group("/auth", authRL.Middleware)
	auth.POST("/register", app.HandleRegister)
	auth.POST("/login", app.HandleLogin)
	auth.POST("/refresh", app.HandleRefresh)

	// Protected routes:

	items := r.Group("/items", RL.Middleware, app.JWTMiddleware)
	// user or higher:
	items.GET("", app.HandleGetItems)
	items.GET("/:uuid", app.HandleGetSingleItem)
	// admin only:
	items.POST("", app.HandleCreateItem)
	items.PATCH("/:uuid", app.HandlePatchItem)
	items.DELETE("/:uuid", app.HandleDeleteItem)

	transactions := r.Group("/transactions", app.JWTMiddleware)
	// user or higher
	transactions.GET("", app.HandleGetAllTransactions)
	transactions.GET("/:uuid", app.HandleGetTransaction)
	// stocker or higher
	transactions.POST("", app.HandleCreateTransaction)

	// RUN:
	if err := r.Start(os.Getenv("SERVER_LISTEN_ADDR")); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
