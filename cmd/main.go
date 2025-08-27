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
	r.POST("/auth", ping)
	r.POST("/login", ping)
	// Protected routes:
	//TODO: whatever

	if err := r.Start(os.Getenv("WAREHOUSE_SERVER_ADDR")); err != nil {
		logger.Error("server error", zap.Error(err))
	}
}

func ping(ctx echo.Context) error {
	ctx.JSON(200, "pong")
	return nil
}
