package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type UserRepo struct {
	DB *sql.DB
}

func (r *UserRepo) GetUserById(ctx context.Context, id int) (*models.User, error) {

	var user_result models.User

	query := `SELECT
		id,
		username,
		firstname,
		lastname,
		pronouns,
		join_date,
		setup_complete,
		archived,
		notes,
		admin,
		force_archive,
		card_tag
		FROM users WHERE users.id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user_result.Id,
		&user_result.Username,
		&user_result.Firstname,
		&user_result.Lastname,
		&user_result.Pronouns,
		&user_result.JoinDate,
		&user_result.SetupComplete,
		&user_result.Archived,
		&user_result.Notes,
		&user_result.Admin,
		&user_result.ForceArchive,
		&user_result.CardTag,
	)

	if err != nil {
		return nil, err
	}

	return &user_result, nil

}
