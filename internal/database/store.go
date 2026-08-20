package database

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"make-backend/internal/database/repos"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

type Store struct {
	Users       repos.UserRepository
	Groups      repos.GroupRepository
	Makerspaces repos.MakerspaceRepository
	Zones       repos.ZoneRepository
	Equipment   repos.EquipmentRepository
	Devices     repos.DeviceRepository
	AuditLogs   repos.AuditLogRepository
}

func NewStore(db *sql.DB) *Store {
	s := &Store{
		Users:       &repos.UserRepo{DB: db},
		Groups:      &repos.GroupRepo{DB: db},
		Makerspaces: &repos.MakerspaceRepo{DB: db},
		Zones:       &repos.ZoneRepo{DB: db},
		Equipment:   &repos.EquipmentRepo{DB: db},
		Devices:     &repos.DeviceRepo{DB: db},
		AuditLogs:   &repos.AuditLogRepo{DB: db},
	}
	fillTestData(context.Background(), slog.Default(), s)
	return s
}
