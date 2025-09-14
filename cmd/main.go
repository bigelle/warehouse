package main

import (
	"context"
	"os"

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
	logger, _ := zap.NewProduction()
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
		middleware.AddTrailingSlash(),
	)

	// Unprotected routes:
	r.POST("/auth/register", app.HandleRegister)
	r.POST("/auth/login", app.HandleLogin)
	r.POST("/auth/refresh", app.HandleRefresh)

	// Protected routes:

	items := r.Group("/items", app.JWTMiddleware)
	// user or higher:
	items.GET("/", app.HandleGetItems)
	items.GET("/:uuid", app.HandleGetSingleItem)
	items.GET("/:uuid/transactions/", ping) // TODO: view all transactions for item
	// admin only:
	items.POST("/", app.HandleCreateItem)
	items.PATCH("/:uuid", app.HandlePatchItem)
	items.DELETE("/:uuid", app.HandleDeleteItem) // TODO: delete item

	transactions := r.Group("/transactions", app.JWTMiddleware)
	// user or higher
	transactions.GET("/", ping)      // TODO: view all transactions
	transactions.GET("/:uuid", ping) // TODO: view specific one
	// stocker or higher
	transactions.POST("/", app.HandleCreateTransaction) // TODO: create one (restock or withdraw)

	// RUN:
	if err := r.Start(os.Getenv("SERVER_LISTEN_ADDR")); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}

func ping(ctx echo.Context) error {
	ctx.JSON(200, "pong")
	return nil
}
