package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/bigelle/warehouse/internal/database"
	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	NotEnoughItemsMessage = "attempt to output a quantity of items exceeding their actual quantity"
)

func (app App) HandleCreateTransaction(c echo.Context) error {
	if !IsAppropriateRole(c.Get("userRole"), schemas.RoleStocker) {
		return echo.ErrForbidden
	}

	uuidStr, ok := c.Get("userID").(string)
	if uuidStr == "" || !ok {
		return echo.ErrForbidden
	}

	uuid, err := UUIDFromString(uuidStr)
	if err != nil {
		return echo.ErrForbidden
	}

	var req schemas.CreateTransactionRequest
	if err = c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	itemUUID, err := UUIDFromString(req.ItemUUID)
	if err != nil {
		return echo.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), TimeoutDatabase*3) // 3 as in 3 requests per TX
	defer cancel()
	tx, err := app.DB.Conn.Begin(ctx)
	if err != nil {
		app.Logger.Error("error starting transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)
	q := app.DB.Queries.WithTx(tx)

	qty, err := q.GetItemQuantity(ctx, itemUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.ErrNotFound
		}
		app.Logger.Error("error getting quantity", zap.Error(err))
		return err
	}

	var tr database.CreateNewTransactionRow
	switch req.Type {
	case schemas.TransactionTypeRestock:
		err = q.SetItemQuantity(ctx, database.SetItemQuantityParams{
			Uuid:     itemUUID,
			Quantity: qty.Quantity + int32(req.Amount),
		})
		if err != nil {
			tx.Rollback(ctx)
			app.Logger.Error("error setting quantity", zap.Error(err))
			return err
		}
		tr, err = q.CreateNewTransaction(ctx, database.CreateNewTransactionParams{
			UserID: uuid,
			ItemID: itemUUID,
			Type:   string(req.Type),
			Amount: int32(req.Amount),
			Status: string(schemas.TransactionStatusSucceeded),
		})
		if err != nil {
			tx.Rollback(ctx)
			app.Logger.Error("error creating transaction", zap.Error(err))
			return err
		}
	case schemas.TransactionTypeWithdraw:
		if qty.Quantity-int32(req.Amount) < 0 {
			msg := NotEnoughItemsMessage
			tr, err = q.CreateNewTransaction(ctx, database.CreateNewTransactionParams{
				UserID: uuid,
				ItemID: itemUUID,
				Type:   string(req.Type),
				Amount: int32(req.Amount),
				Status: string(schemas.TransactionStatusFailed),
				Reason: &msg,
			})
			if err != nil {
				tx.Rollback(ctx)
				app.Logger.Error("error creating transaction", zap.Error(err))
				return err
			}
			// FIXME: send the reason in json response
			return echo.NewHTTPError(http.StatusBadRequest, msg)
		}
		err := q.SetItemQuantity(ctx, database.SetItemQuantityParams{
			Uuid:     itemUUID,
			Quantity: qty.Quantity + int32(req.Amount),
		})
		if err != nil {
			tx.Rollback(ctx)
			app.Logger.Error("error setting quantity", zap.Error(err))
			return err
		}
		tr, err = q.CreateNewTransaction(ctx, database.CreateNewTransactionParams{
			UserID: uuid,
			ItemID: itemUUID,
			Type:   string(req.Type),
			Amount: int32(req.Amount),
			Status: string(schemas.TransactionStatusSucceeded),
		})
		if err != nil {
			tx.Rollback(ctx)
			app.Logger.Error("error creating transaction", zap.Error(err))
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err // what do I do here?
	}

	return c.JSON(http.StatusAccepted, schemas.CreateTransactionResponse{
		Type:            req.Type,
		ItemUUID:        req.ItemUUID,
		TransactionUUID: tr.ID.String(),
		Amount:          req.Amount,
		Status:          schemas.TransactionStatusSucceeded,
		CreatedAt:       tr.CreatedAt.Time.Unix(),
	})
}
