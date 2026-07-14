package logging

import "make-backend/internal/database"

type Logger struct {
	store    *database.Store
	AuditLog *AuditLogger
}

func NewLogger(store *database.Store) *Logger {
	return &Logger{
		store:    store,
		AuditLog: &AuditLogger{store: store},
	}
}
