package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type ZoneRepository interface {
	GetZoneById(ctx context.Context, id int) (*models.Zone, error)
	GetZonesByMakerspaceId(ctx context.Context, makerspaceId int) ([]*models.Zone, error)
}

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

func (r *ZoneRepo) GetZonesByMakerspaceId(ctx context.Context, makerspaceId int) ([]*models.Zone, error) {
	query := "SELECT id, makerspace_id, name, hidden FROM zones WHERE zones.makerspace_id = $1"

	rows, err := r.DB.QueryContext(ctx, query, makerspaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*models.Zone

	for rows.Next() {
		var zone models.Zone

		err := rows.Scan(&zone.Id, &zone.MakerspaceId, &zone.Name, &zone.Hidden)
		if err != nil {
			return nil, err
		}

		zones = append(zones, &zone)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return zones, err
}
