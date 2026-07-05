package auth

import (
	"database/sql"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

func SetupSessionManager(db *sql.DB) *scs.SessionManager {
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 48 * time.Hour

	return sessionManager
}
