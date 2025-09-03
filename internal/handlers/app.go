package handlers

import (
	"github.com/bigelle/warehouse/internal/database"
	"go.uber.org/zap"
)

type App struct {
	Database *database.Queries
	// Cache
	// TODO: maybe use interface
	Logger *zap.Logger
	Config Config
}

//TODO: maybe some shorteners for logging or shi

type Config struct {
	JWTAccessSecret  []byte
	JWTRefreshSecret []byte
}
