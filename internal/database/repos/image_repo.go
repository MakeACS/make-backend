package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type ImageRepository interface {
	GetImageById(ctx context.Context, id int) (*models.Image, error)
}

type ImageRepo struct {
	DB *sql.DB
}

func (r *ImageRepo) GetImageById(ctx context.Context, id int) (*models.Image, error) {
	var image_result models.Image

	query := "SELECT id, identifier FROM images WHERE images.id = $1"

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&image_result.Id,
		&image_result.Identifier,
	)

	if err != nil {
		return nil, err
	}

	return &image_result, nil
}
