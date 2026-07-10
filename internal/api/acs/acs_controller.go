package acs

import "make-backend/internal/database/models"

type ACSController interface {
	sendCoreAuthToResponse(core models.AccessDevice, response ServerAuthToResponse) bool
	sendCoreConfigUpdate(core models.AccessDevice, update ServerConfigUpdateRequest) bool
	sendCoreInfoResponse(core models.AccessDevice, response ServerInfoResponse) bool
	sendCoreCommand(core models.AccessDevice, command ServerCommand) bool
	sendWelcomeResponse(device models.Device, response WelcomeResponse) bool
	getName() string
}
