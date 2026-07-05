package database

import (
	"database/sql"
	"make-backend/internal/database/repos"
)

type Store struct {
	Users       repos.UserRepository
	Makerspaces repos.MakerspaceRepository
	Zones       repos.ZoneRepository
	Equipment   repos.EquipmentRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Users:       &repos.UserRepo{DB: db},
		Makerspaces: &repos.MakerspaceRepo{DB: db},
		Zones:       &repos.ZoneRepo{DB: db},
		Equipment:   &repos.EquipmentRepo{DB: db},
	}
}
