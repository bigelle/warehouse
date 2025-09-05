package handlers

import (
	"context"
	"errors"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

func (app App) HandleGetItems(c echo.Context) error {
	if role, ok := c.Get("userRole").(schemas.Role); !ok || role < schemas.RoleUser {
		return echo.ErrForbidden
	}

	var req schemas.GetItemsRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	if req.Limit == nil {
		req.Limit = &schemas.GetItemsRequestDefaultLimit
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase)
	defer cancel()
	found, err := app.Database.GetNItemsOffset(ctx, database.GetNItemsOffsetParams{
		Limit:  int32(*req.Limit),
		Offset: int32(req.Offset),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrNotFound
		}
		return err
	}

	nFound := len(found)
	if nFound == 0 {
		return echo.ErrNotFound
	}

	items := make([]schemas.Item, nFound)
	for i := range nFound {
		items[i] = schemas.Item{
			ID:       found[i].Uuid.String(),
			Name:     found[i].Name,
			Quantity: int(found[i].Quantity),
		}
	}
	return c.JSON(200, schemas.GetItemsResponse{
		NResults: nFound,
		Items:    items,
	})
}
