package database

import (
	"database/sql"
	"embed"
	"make-backend/internal/database/repos"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

type Store struct {
	Users       repos.UserRepository
	Makerspaces repos.MakerspaceRepository
	Zones       repos.ZoneRepository
	Equipment   repos.EquipmentRepository
	Devices     repos.DeviceRepository
	AuditLogs   repos.AuditLogRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Users:       &repos.UserRepo{DB: db},
		Makerspaces: &repos.MakerspaceRepo{DB: db},
		Zones:       &repos.ZoneRepo{DB: db},
		Equipment:   &repos.EquipmentRepo{DB: db},
		Devices:     &repos.DeviceRepo{DB: db},
		AuditLogs:   &repos.AuditLogRepo{DB: db},
	}
}
