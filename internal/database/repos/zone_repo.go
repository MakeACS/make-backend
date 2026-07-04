package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type ZoneRepo struct {
	DB *sql.DB
}

func (r *ZoneRepo) GetZoneById(ctx context.Context, id int) (*models.Zone, error) {
	var zone_result models.Zone

	query := `SELECT
		id,
		makerspace_id,
		name,
		hidden
		FROM zones WHERE zones.id = $1
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&zone_result.Id,
		&zone_result.MakerspaceId,
		&zone_result.Name,
		&zone_result.Hidden,
	)

	if err != nil {
		return nil, err
	}

	return &zone_result, nil
}
