package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id int) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, email string) (int, error)
	IsTrainer(ctx context.Context, id int) (bool, error)
	IsTrainerFor(ctx context.Context, user_id int, equipment_id int) (bool, error)
}

type UserRepo struct {
	DB *sql.DB
}

func (r *UserRepo) GetUserById(ctx context.Context, id int) (*models.User, error) {
	var user_result models.User

	query := `SELECT
		id,
		email,
		full_name,
		preferred_name,
		pronouns,
		join_date,
		setup_complete,
		archived,
		notes,
		force_archive
		FROM users WHERE users.id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user_result.Id,
		&user_result.Email,
		&user_result.FullName,
		&user_result.PreferredName,
		&user_result.Pronouns,
		&user_result.JoinDate,
		&user_result.SetupComplete,
		&user_result.Archived,
		&user_result.Notes,
		&user_result.ForceArchive,
	)

	if err != nil {
		return nil, err
	}

	return &user_result, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user_result models.User

	query := `SELECT
		id,
		email,
		full_name,
		preferred_name,
		pronouns,
		join_date,
		setup_complete,
		archived,
		notes,
		force_archive
		FROM users WHERE users.email = $1`

	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user_result.Id,
		&user_result.Email,
		&user_result.FullName,
		&user_result.PreferredName,
		&user_result.Pronouns,
		&user_result.JoinDate,
		&user_result.SetupComplete,
		&user_result.Archived,
		&user_result.Notes,
		&user_result.ForceArchive,
	)

	if err != nil {
		return nil, err
	}

	return &user_result, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, email string) (int, error) {
	var new_id int

	query := `INSERT INTO users (email) VALUES ($1) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, email).Scan(&new_id)
	if err != nil {
		return 0, err
	}

	return new_id, nil
}

func (r *UserRepo) IsTrainer(ctx context.Context, id int) (bool, error) {
	var isTrainer bool

	query := `SELECT EXISTS(SELECT 1 FROM trainers WHERE user_id = $1)`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(&isTrainer)
	if err != nil {
		return false, err
	}

	return isTrainer, nil
}

func (r *UserRepo) IsTrainerFor(ctx context.Context, user_id int, equipment_id int) (bool, error) {
	var isTrainer bool

	query := `SELECT EXISTS(SELECT 1 FROM trainers WHERE user_id = $1 AND equipment_id = $2)`

	err := r.DB.QueryRowContext(ctx, query, user_id, equipment_id).Scan(&isTrainer)
	if err != nil {
		return false, err
	}

	return isTrainer, nil
}
