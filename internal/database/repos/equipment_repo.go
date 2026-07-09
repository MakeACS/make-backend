package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type EquipmentRepository interface {
	GetEquipmentById(ctx context.Context, id int) (*models.Equipment, error)
}

type EquipmentRepo struct {
	DB *sql.DB
}

func (r *EquipmentRepo) GetEquipmentById(ctx context.Context, id int) (*models.Equipment, error) {
	var equipment_result models.Equipment

	query := `SELECT
		id,
		name,
		sub_name,
		zone_id,
		hidden,
		image_id,
		sop_url,
		sign_off_url,
		description,
		reservable,
		reservation_only,
		reservation_instructions,
		needs_welcome,
		requires_in_person,
		requires_trainer
		FROM equipment WHERE equipment.id = $1
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&equipment_result.Id,
		&equipment_result.Name,
		&equipment_result.SubName,
		&equipment_result.ZoneId,
		&equipment_result.Hidden,
		&equipment_result.ImageId,
		&equipment_result.SopUrl,
		&equipment_result.SignOffUrl,
		&equipment_result.Description,
		&equipment_result.Reservable,
		&equipment_result.ReservationOnly,
		&equipment_result.ReservationInstructions,
		&equipment_result.NeedsWelcome,
		&equipment_result.RequiresInPerson,
		&equipment_result.RequiresTrainer,
	)
	if err != nil {
		return nil, err
	}

	return &equipment_result, nil
}
