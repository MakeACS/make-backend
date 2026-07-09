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

func (r *DeviceRepo) GetAccessDeviceById(ctx context.Context, id int) (*models.AccessDevice, error) {
	var device_result models.AccessDevice

	query := `SELECT
		device_id,
		channels,
		temp_duration,
		current_card_tag,
		last_status,
		session_start,
		flags,
		sealed_deployment,
		reported_deployment
		FROM access_devices WHERE device_id = $1
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&device_result.DeviceId,
		&device_result.Channels,
		&device_result.TempDuration,
		&device_result.CurrentCardTag,
		&device_result.LastStatus,
		&device_result.SessionStart,
		&device_result.Flags,
		&device_result.SealedDeployment,
		&device_result.ReportedDeployment,
	)
	if err != nil {
		return nil, err
	}

	return &device_result, nil
}

func (r *DeviceRepo) GetAccessChannelByDeviceAndChannelId(ctx context.Context, device_id int, channel_id int) (*models.AccessChannel, error) {
	var access_channel_result models.AccessChannel

	query := `SELECT id, device_id, channel_id, state, temp_duration FROM access_channels WHERE device_id = $1 AND channel_id = $2`

	err := r.DB.QueryRowContext(ctx, query, device_id, channel_id).Scan(
		&access_channel_result.Id,
		&access_channel_result.DeviceId,
		&access_channel_result.ChannelId,
		&access_channel_result.State,
		&access_channel_result.TempDuration,
	)
	if err != nil {
		return nil, err
	}

	return &access_channel_result, nil
}
