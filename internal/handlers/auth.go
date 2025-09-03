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

		// preventing duplicates
		ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
		defer cancel()
		_, err := app.Database.GetUserByUsername(ctx, req.Username)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				app.Logger.Error("finding user", zap.Error(err))
				return err
			}
		}

		// hashing
		hash, err := HashPassword(req.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		// creating user
		cancel() // JUST IN CASE
		ctx, cancel = context.WithTimeout(c.Request().Context(), TimeoutDatabase)
		defer cancel()
		usr, err := app.Database.CreateUser(ctx,
			database.CreateUserParams{
				Username:     req.Username,
				PasswordHash: hash,
				Role:         schemas.RoleUser.String(),
			},
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505":
					return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
				case "23514":
					return echo.NewHTTPError(http.StatusBadRequest, "invalid role type")
				case "23502":
					return echo.NewHTTPError(http.StatusBadRequest, "null value")
				default:
					app.Logger.Error("unexpected database error while creating user", zap.String("code", pgErr.Code))
					return echo.NewHTTPError(http.StatusInternalServerError) // 500 for now, maybe I'll leave it like that
				}
			}
			app.Logger.Error("unexpected database error while creating user", zap.String("code", pgErr.Code))
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		return c.JSON(200, schemas.RegisterResponse{
			Name: usr.Username,
			Role: schemas.RoleFromString(usr.Role),
		})

	}
}

func HandleLogin(app *App) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req schemas.LoginRequest
		if err := c.Bind(&req); err != nil {
			app.Logger.Error("binding error", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, "bad request")
		}

		//TODO: validating things

		ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
		defer cancel()
		usr, err := app.Database.GetUserByUsername(ctx, req.Username)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return echo.NewHTTPError(http.StatusUnauthorized, "wrong username or password")
			}
			return err
		}

		if !IsCorrectPassword(req.Password, usr.PasswordHash) {
			return echo.NewHTTPError(http.StatusUnauthorized, "wrong username or password")
		}

		access, err := GenerateAccessJWT(usr.Username, usr.Role, app.Config.JWTAccessSecret, 15*time.Minute)
		if err != nil {
			return err // nothing i can do
		}
		refresh, err := GenerateRefreshJWT(usr.Username, app.Config.JWTRefreshSecret, 7*24*time.Hour)
		if err != nil {
			return err
		}

		// NOTE: maybe i can do it in parallel
		cancel()
		ctx, cancel = context.WithTimeout(c.Request().Context(), TimeoutDatabase)
		defer cancel()
		_, err = app.Database.SetRefreshToken(ctx, database.SetRefreshTokenParams{
			RefreshToken: refresh,
			ID:           usr.ID,
		})
		if err != nil {
			// idk, it's not as bad, the user is logged in until access token expires,
			// but can't refresh it so he will login again
			if errors.Is(err, pgx.ErrNoRows) {
				app.Logger.Error("setting refresh token for non-existing user", zap.String("id", usr.ID.String()))
			} else {
				app.Logger.Error("unexpected error while setting refresh token", zap.Error(err))
			}
		}

		c.SetCookie(&http.Cookie{
			Name:     "refresh-token",
			Value:    refresh,
			Path:     "/refresh",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
		})

		return c.JSON(200, schemas.LoginResponse{
			AccessToken: access,
		})
	}
}
