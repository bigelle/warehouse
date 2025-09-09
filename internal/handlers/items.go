package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (app App) HandleCreateItem(c echo.Context) error {
	if !IsAppropriateRole(c.Get("userRole"), schemas.RoleAdmin) {
		return echo.ErrForbidden
	}

	var req schemas.CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	item, err := app.Database.CreateItem(ctx, req.Name)
	if err != nil {
		var uniqueErr *pgconn.PgError
		if ok := errors.As(err, &uniqueErr); ok && uniqueErr.Code == "23505" {
			return echo.NewHTTPError(http.StatusConflict, "item with this name already exists")
		}
		return err
	}

	return c.JSON(200, schemas.CreateItemResponse{
		UUID:      item.Uuid.String(),
		Name:      item.Name,
		CreatedAt: item.CreatedAt.Time.Unix(),
	})
}

func (app App) HandleGetItems(c echo.Context) error {
	if !IsAppropriateRole(c.Get("userRole"), schemas.RoleUser) {
		return echo.ErrForbidden
	}

	var req schemas.GetItemsRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	if req.Limit == 0 {
		req.Limit = schemas.GetItemsRequestDefaultLimit
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	found, err := app.Database.GetNItemsOffset(ctx, database.GetNItemsOffsetParams{
		Limit:  int32(req.Limit),
		Offset: int32(req.Offset),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrNotFound
		}
		app.Logger.Error("getting rows", zap.Error(err))
		return err
	}

	nFound := len(found)
	if nFound == 0 {
		return echo.ErrNotFound
	}

	items := make([]schemas.Item, nFound)
	for i := range nFound {
		items[i] = schemas.Item{
			UUID:     found[i].Uuid.String(),
			Name:     found[i].Name,
			Quantity: int(found[i].Quantity),
		}
	}
	return c.JSON(200, schemas.GetItemsResponse{
		NResults: nFound,
		Items:    items,
	})
}

func (app App) HandleGetSingleItem(c echo.Context) error {
	if !IsAppropriateRole(c.Get("userRole"), schemas.RoleUser) {
		return echo.ErrForbidden
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		return echo.ErrBadRequest
	}

	strUuid, err := UUIDFromString(uuid)
	if err != nil {
		return echo.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	item, err := app.Database.GetItem(ctx, strUuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrNotFound
		}
		return err
	}

	return c.JSON(200, schemas.Item{
		UUID:     item.Uuid.String(),
		Name:     item.Name,
		Quantity: int(item.Quantity),
	})
}

func (app App) HandlePatchItem(c echo.Context) error {
	if !IsAppropriateRole(c.Get("userRole"), schemas.RoleAdmin) {
		return echo.ErrForbidden
	}

	strUUID := c.Param("uuid")
	if strUUID == "" {
		return echo.ErrBadRequest
	}

	var req schemas.PatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	uuid, err := UUIDFromString(strUUID)
	if err != nil {
		return echo.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	item, err := app.Database.PatchItem(ctx, database.PatchItemParams{
		Uuid:     uuid,
		Name:     req.Name,
		Quantity: req.Quantity,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrNotFound
		}
		return err
	}

	return c.JSON(200, schemas.Item{
		UUID:     item.Uuid.String(),
		Name:     item.Name,
		Quantity: int(item.Quantity),
	})
}

func (app App) HandleDeleteItem(c echo.Context) error {
	if !IsAppropriateRole(c.Get("userRole"), schemas.RoleAdmin) {
		return echo.ErrForbidden
	}

	strUUID := c.Param("uuid")
	if strUUID == "" {
		return echo.ErrBadRequest
	}

	uuid, err := UUIDFromString(strUUID)
	if err != nil {
		return echo.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	err = app.Database.DeleteItem(ctx, uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrNotFound
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
