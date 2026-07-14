package resolvers

import (
	"make-backend/internal/database"
	"make-backend/internal/logging"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Store  *database.Store
	Logger *logging.Logger
}
