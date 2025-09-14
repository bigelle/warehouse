package handlers

import (
	"github.com/bigelle/warehouse/internal/database"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type App struct {
	DB     Database
	Logger *zap.Logger
	Config Config
}

type Database struct {
	Queries *database.Queries
	Conn    *pgx.Conn
}

type Config struct {
	JWTAccessSecret  []byte
	JWTRefreshSecret []byte
}
