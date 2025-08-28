package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load("app.env"); err != nil {
		logger.Error("failed to load environment variables", zap.Error(err))
	}

	r := echo.New()
	// Unprotected routes:
	r.POST("/auth/register", ping)
	r.POST("/auth/login", ping)

	// Protected routes:
	// any authorized user:
	r.GET("/items", ping)     // get all items, may accept offset
	r.GET("/items/:id", ping) // get a specific item
	r.GET("/notice", ping)    // see if we run out of something.  TODO: rename it
	// admin only:
	r.POST("/items", ping)       // add a new item to tracking
	r.PATCH("/items/:id", ping)  // change qty, description, etc.
	r.DELETE("/items/:id", ping) // delete an item from tracking

	if err := r.Start(os.Getenv("WAREHOUSE_SERVER_ADDR")); err != nil {
		logger.Error("server error", zap.Error(err))
	}
}

func ping(ctx echo.Context) error {
	ctx.JSON(200, "pong")
	return nil
}
