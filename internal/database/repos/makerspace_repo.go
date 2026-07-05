package repos

import (
	"context"
	"database/sql"
	"fmt"
	"make-backend/internal/database/models"
)

type MakerspaceRepository interface {
	GetMakerspaceById(ctx context.Context, id int) (*models.Makerspace, error)
	CreateMakerspace(ctx context.Context, name string, hidden bool) (int, error)
	AddManager(ctx context.Context, makerspace_id int, user_id int) error
	AddStaff(ctx context.Context, makerspace_id int, user_id int) error
}

type MakerspaceRepo struct {
	DB *sql.DB
}

func (r *MakerspaceRepo) GetMakerspaceById(ctx context.Context, id int) (*models.Makerspace, error) {
	var makerspace_result models.Makerspace

	query := `SELECT
		id,
		name,
		subtitle,
		description,
		docs_url,
		image_id,
		hidden
		FROM makerspaces where makerspaces.id = $1
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&makerspace_result.Id,
		&makerspace_result.Name,
		&makerspace_result.Subtitle,
		&makerspace_result.Description,
		&makerspace_result.DocsUrl,
		&makerspace_result.ImageId,
		&makerspace_result.Hidden,
	)

	if err != nil {
		return nil, err
	}

	return &makerspace_result, nil
}

func (r *MakerspaceRepo) CreateMakerspace(ctx context.Context, name string, hidden bool) (int, error) {
	var id_result int

	query := `INSERT INTO makerspaces (name, hidden) VALUES ($1, $2) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, name, hidden).Scan(&id_result)
	if err != nil {
		return 0, err
	}

	return id_result, nil
}

func (r *MakerspaceRepo) AddManager(ctx context.Context, makerspace_id int, user_id int) error {
	query := `INSERT INTO managers (makerspace_id, user_id) VALUES ($1, $2)`

	_, err := r.DB.ExecContext(ctx, query, makerspace_id, user_id)
	if err != nil {
		return fmt.Errorf("Failed to insert manager (makerspace: %d, user: %d): %w", makerspace_id, user_id, err)
	}

	return nil
}

func (r *MakerspaceRepo) AddStaff(ctx context.Context, makerspace_id int, user_id int) error {
	query := `INSERT INTO staff (makerspace_id, user_id) VALUES ($1, $2)`

	_, err := r.DB.ExecContext(ctx, query, makerspace_id, user_id)
	if err != nil {
		return fmt.Errorf("Failed to insert staff (makerspace: %d, user: %d): %w", makerspace_id, user_id, err)
	}

	return nil
}

func (r *MakerspaceRepo) GetManagers(ctx context.Context, makerspace_id int) ([]*models.User, error) {
	query := `SELECT
		users.id
		users.username
		users.firstname
		users.lastname
		users.pronouns
		users.join_date
		users.setup_complete
		users.archived
		users.notes
		users.admin
		users.force_archive
		users.card_tag
		FROM users JOIN managers ON users.id = managers.user_id
		WHERE managers.makerspace_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, makerspace_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.Id,
			&user.Username,
			&user.Firstname,
			&user.Lastname,
			&user.Pronouns,
			&user.JoinDate,
			&user.SetupComplete,
			&user.Archived,
			&user.Notes,
			&user.Admin,
			&user.ForceArchive,
			&user.CardTag,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
