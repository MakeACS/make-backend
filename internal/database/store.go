package database

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
	"make-backend/internal/database/repos"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id int) (*models.User, error)
}

type MakerspaceRepository interface {
	GetMakerspaceById(ctx context.Context, id int) (*models.Makerspace, error)
}

type Store struct {
	Users       UserRepository
	Makerspaces MakerspaceRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Users:       &repos.UserRepo{DB: db},
		Makerspaces: &repos.MakerspaceRepo{DB: db},
	}
}
