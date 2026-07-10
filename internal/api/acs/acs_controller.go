package acs

import "make-backend/internal/database/models"

type ACSController interface {
	SendAccessDeviceAuthToResponse(adev models.AccessDevice, response ServerAuthToResponse) bool
	SendAccessDeviceConfigUpdate(adev models.AccessDevice, update ServerConfigUpdateRequest) bool
	SendAccessDeviceInfoResponse(adev models.AccessDevice, response ServerInfoResponse) bool
	SendAccessDeviceCommand(adev models.AccessDevice, command ServerCommand) bool
	SendWelcomeResponse(device models.Device, response WelcomeResponse) bool
	GetName() string
}
