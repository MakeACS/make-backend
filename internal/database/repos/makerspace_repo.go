package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type MakerspaceRepository interface {
	GetMakerspaceById(ctx context.Context, id int) (*models.Makerspace, error)
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
