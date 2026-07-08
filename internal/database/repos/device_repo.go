package repos

import (
	"context"
	"database/sql"
	"make-backend/internal/database/models"
)

type DeviceRepository interface {
	GetDeviceById(ctx context.Context, id int) (*models.Device, error)
	GetDeviceBySN(ctx context.Context, sn string) (*models.Device, error)
}

type DeviceRepo struct {
	DB *sql.DB
}

func (r *DeviceRepo) GetDeviceById(ctx context.Context, id int) (*models.Device, error) {
	var equipment_result models.Device

	query := `SELECT
		id,
		name,
		sn,
		paired,
		hardware,
		firmware,
		target_firmware,
		key_cycle,
		makerspace_id
		FROM devices WHERE devices.id = $1
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&equipment_result.Id,
		&equipment_result.Name,
		&equipment_result.SN,
		&equipment_result.Paired,
		&equipment_result.Hardware,
		&equipment_result.Firmware,
		&equipment_result.TargetFirmware,
		&equipment_result.KeyCycle,
		&equipment_result.MakerspaceId,
	)
	if err != nil {
		return nil, err
	}

	return &equipment_result, nil
}

func (r *DeviceRepo) GetDeviceBySN(ctx context.Context, sn string) (*models.Device, error) {
	var equipment_result models.Device

	query := `SELECT
		id,
		name,
		sn,
		paired,
		hardware,
		firmware,
		target_firmware,
		key_cycle,
		makerspace_id
		FROM devices WHERE devices.sn = $1
	`

	err := r.DB.QueryRowContext(ctx, query, sn).Scan(
		&equipment_result.Id,
		&equipment_result.Name,
		&equipment_result.SN,
		&equipment_result.Paired,
		&equipment_result.Hardware,
		&equipment_result.Firmware,
		&equipment_result.TargetFirmware,
		&equipment_result.KeyCycle,
		&equipment_result.MakerspaceId,
	)
	if err != nil {
		return nil, err
	}

	return &equipment_result, nil
}
