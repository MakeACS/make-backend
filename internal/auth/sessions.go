package auth

import (
	"database/sql"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

var sessionManager *scs.SessionManager

func SetupSessionManager(db *sql.DB) {
	sessionManager = scs.New()
	sessionManager.Store = postgresstore.New(db)
}
