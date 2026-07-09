package auth

import (
	"database/sql"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

func SetupSessions(db *sql.DB) *scs.SessionManager {

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)

	return sessionManager
}
