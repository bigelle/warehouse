package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	TimeoutDatabase = 500 * time.Millisecond
)

func (app App) HandleRegister(c echo.Context) error {
	var req schemas.RegisterRequest
	if err := c.Bind(&req); err != nil {
		app.Logger.Error("binding error", zap.Error(err))
		return echo.ErrBadRequest
	}
	// TODO: validating things

	// hashing
	hash, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	// creating user
	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	usr, err := app.DB.Queries.CreateUser(ctx,
		database.CreateUserParams{
			Username:     req.Username,
			PasswordHash: hash,
			Role:         req.Role.String(),
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return echo.ErrConflict
			case "23514":
				return echo.NewHTTPError(http.StatusBadRequest, "invalid role type")
			case "23502":
				return echo.NewHTTPError(http.StatusBadRequest, "null value")
			default:
				app.Logger.Error("unexpected database error while creating user", zap.String("code", pgErr.Code))
				return err // 500 for now, maybe I'll leave it like that
			}
		}
		app.Logger.Error("unexpected database error while creating user", zap.String("code", pgErr.Code))
		return err
	}

	return c.JSON(200, schemas.RegisterResponse{
		Username: usr.Username,
		UUID:     usr.ID.String(),
		Role:     schemas.RoleFromString(usr.Role),
	})
}

func (app App) HandleLogin(c echo.Context) error {
	var req schemas.LoginRequest
	if err := c.Bind(&req); err != nil {
		app.Logger.Error("binding error", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	// TODO: validating things

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	usr, err := app.DB.Queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusUnauthorized, "wrong username or password")
		}
		return err
	}

	if !IsCorrectPassword(req.Password, usr.PasswordHash) {
		return echo.NewHTTPError(http.StatusUnauthorized, "wrong username or password")
	}

	access, err := GenerateAccessJWT(usr.ID.String(), usr.Role, app.Config.JWTAccessSecret, 15*time.Minute)
	if err != nil {
		return err // nothing i can do
	}
	refresh, err := GenerateRefreshJWT(usr.ID.String(), app.Config.JWTRefreshSecret, 7*24*time.Hour)
	if err != nil {
		return err
	}

	// NOTE: maybe i can do it in parallel
	cancel()
	ctx, cancel = context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	_, err = app.DB.Queries.SetRefreshToken(ctx, database.SetRefreshTokenParams{
		RefreshToken: refresh,
		ID:           usr.ID,
	})
	if err != nil {
		// idk, it's not as bad, the user is logged in until access token expires,
		// but can't refresh it so he will login again
		if errors.Is(err, pgx.ErrNoRows) {
			app.Logger.Error("setting refresh token for non-existing user")
		} else {
			app.Logger.Error("unexpected error while setting refresh token", zap.Error(err))
		}
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh",
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

func (app App) HandleRefresh(c echo.Context) error {
	refresh, err := c.Cookie("refresh")
	if err != nil {
		return echo.ErrUnauthorized
	}

	// Validating refresher:
	token, err := jwt.Parse(refresh.Value, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.ErrUnauthorized
		}
		return app.Config.JWTRefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return echo.ErrUnauthorized
	}
	refreshClaims := token.Claims.(jwt.MapClaims)

	// Getting uuid:
	subj, err := refreshClaims.GetSubject()
	if err != nil {
		return err
	}
	uuid, err := UUIDFromString(subj)
	if err != nil {
		return err
	}

	// Generating access:
	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	usrRole, err := app.DB.Queries.GetUserRole(ctx, uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrUnauthorized
		}
		return err
	}
	access, err := GenerateAccessJWT(usrRole.ID.String(), usrRole.Role, app.Config.JWTAccessSecret, 15*time.Minute)
	if err != nil {
		return err
	}

	return c.JSON(200, schemas.LoginResponse{
		AccessToken: access,
	})
}

func (app App) JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth == "" {
			return echo.ErrUnauthorized
		}
		parts := strings.SplitN(auth, " ", 2)

		// testing for bad format:
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.ErrUnauthorized
		}

		// parsing
		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.ErrUnauthorized
			}
			return app.Config.JWTAccessSecret, nil
		})
		if err != nil || !token.Valid {
			return echo.ErrUnauthorized
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return echo.ErrUnauthorized
		}

		// setting
		c.Set("userID", claims["sub"])
		c.Set("userRole", schemas.RoleFromString(claims["role"].(string)))

		return next(c)
	}
}
