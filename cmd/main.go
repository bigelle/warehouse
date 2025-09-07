package main

import (
	"context"
	"os"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/handlers"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
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
	db := database.New(conn)

	// APP:
	app := handlers.App{
		Database: db,
		Logger:   logger,
	}

	// ROUTER:
	r := echo.New()
	// Unprotected routes:
	r.POST("/auth/register", app.HandleRegister)
	r.POST("/auth/login", app.HandleLogin)
	r.POST("/auth/refresh", app.HandleRefresh)
	// Protected routes:
	// any authorized user:
	r.GET("/items", app.HandleGetItems, app.JWTMiddleware) // get all items, may accept offset
	r.GET("/items/:uuid", ping)                            // get a specific item
	r.GET("/notice", ping)                                 // see if we run out of something.  TODO: rename it
	// admin only:
	r.POST("/items", app.HandleCreateItem, app.JWTMiddleware) // add a new item to tracking
	r.PATCH("/items/:id", ping)                               // change qty, description, etc.
	r.DELETE("/items/:id", ping)                              // delete an item from tracking

	// RUN:
	if err := r.Start(os.Getenv("WAREHOUSE_LISTEN_ADDR")); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}

func ping(ctx echo.Context) error {
	ctx.JSON(200, "pong")
	return nil
}
