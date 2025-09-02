package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TODO: move it somewhere else
type App struct {
	Database *database.Queries
	// Cache
	Logger *zap.Logger
}

const (
	TimeoutDatabase = 500 * time.Millisecond
)

func HandleRegister(app *App) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req schemas.RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		//TODO: validating things
		//TODO: looking for cache(?)

		// preventing duplicates
		ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
		defer cancel()
		_, err := app.Database.GetUserByUsername(ctx, req.Name)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				app.Logger.Error("finding user", zap.Error(err))
				return c.NoContent(http.StatusInternalServerError)
			}
		}

		// hashing
		hash, err := HashPassword(req.Password)
		if err != nil {
			// TODO: check if it's even possible
			return c.NoContent(http.StatusBadRequest) // bad request for now since I'm not validating anything rn
		}

		// creating user
		cancel() // JUST IN CASE
		ctx, cancel = context.WithTimeout(c.Request().Context(), TimeoutDatabase)
		defer cancel()
		usr, err := app.Database.CreateUser(ctx,
			database.CreateUserParams{
				Username:     req.Name,
				PasswordHash: hash,
				Role:         string(schemas.AccessLevelUser),
			},
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505":
					return c.String(http.StatusBadRequest, "user already exists")
				case "23514":
					return c.String(http.StatusBadRequest, "invalid role type")
				case "23502":
					return c.String(http.StatusBadRequest, "null value")
				default:
					app.Logger.Error("unexpected database error while creating user", zap.String("code", pgErr.Code))
					return c.NoContent(http.StatusInternalServerError) // 500 for now, maybe I'll leave it like that
				}
			}
			app.Logger.Error("unexpected database error while creating user", zap.String("code", pgErr.Code))
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(200, schemas.RegisterResponse{
			Name: usr.Username,
			//TODO: getter for enum JUST IN CASE
			Role: schemas.AccessLevel(usr.Role),
		})

	}
}
