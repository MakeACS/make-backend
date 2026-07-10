package acs

import "make-backend/internal/database"

type ACSOrchestrator struct {
	store *database.Store
}

func (orc *ACSOrchestrator) RegisterDevice(deviceID int64, controller ACSController) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) GetDeviceController(deviceID int64) *ACSController {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleCoreStatusReport(deviceID int64, statusReport AccessDeviceStatusReport) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleCoreStateChangeReport(deviceID int64, stateChangeReport AccessDeviceStateChangeReport) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleCoreLogRequest(deviceID int64, logRequest AccessDeviceLogRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleCoreAuthToRequest(deviceID int64, authToRequest AccessDeviceAuthToRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) handleCoreInfoRequest(deviceID int64, infoRequest AccessDeviceInfoRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleSendAccessDeviceCommand(deviceID int64, command ServerCommand) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleWelcomeRequest(makerspaceID int64, deviceID int64, cardTagID string) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) CommandAllAccessDevices(command ServerCommand) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) CommandMakerspaceAccessDevices(makerspaceID int64, command ServerCommand) {
	panic("unimplemented")

}
