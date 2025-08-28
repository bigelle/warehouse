package handlers

import "github.com/labstack/echo/v4"

// TODO: move it somewhere else
type App struct {
	// Database
	// Cache
	// Logger
}

func HandleRegister(app *App) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, "pong")
	}
}
