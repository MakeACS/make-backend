package acs

import (
	"context"
	"log/slog"
	"make-backend/internal/database"
	"make-backend/internal/database/models"
	"make-backend/internal/logging"
)

type ACSOrchestrator struct {
	store       *database.Store
	controllers map[int]ACSController
	slog        *slog.Logger
	logger      *logging.Logger
}

func (orc *ACSOrchestrator) RegisterDevice(deviceID int, controller ACSController) {
	orc.controllers[deviceID] = controller
}

func (orc *ACSOrchestrator) GetDeviceController(deviceID int) ACSController {
	val, ok := orc.controllers[deviceID]
	if !ok {
		return nil
	}
	return val
}

func (orc *ACSOrchestrator) HandleAccessDeviceStatusReport(deviceID int, statusReport AccessDeviceStatusReport) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) endSession(dev *models.AccessDevice, oldCardTag string) {
	panic("unimplemented")
}
func (orc *ACSOrchestrator) startSession(dev *models.AccessDevice, cardTag string) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceStateChangeReport(deviceID int, stateChangeReport AccessDeviceStateChangeReport) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceLogRequest(deviceID int, logRequest AccessDeviceLogRequest) {
	dev, err := orc.store.Devices.GetAccessDeviceById(context.TODO(), int(deviceID))
	if err != nil {
		orc.slog.Warn("Couldn't find device for id for log request report", "id", deviceID)
		return
	}
	data := map[string]any{
		"makerspace_id":    dev.MakerspaceId,
		"device_id":        dev.Id,
		"device_label":     dev.Name,
		"original_message": logRequest.Message,
	}
	if logRequest.AuditLog {
		fullMessage := "Message from {device}: " + logRequest.Message
		orc.logger.AuditLog.CreateWithData(dev.MakerspaceId, "base.access_device.log."+logRequest.Category, data, fullMessage, dev.LogEntity())
	} else {
		panic("unimplemented")
	}

}

func HandleAccessDeviceConfigReport(deviceID int, configReport AccessDeviceConfigReport) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceAuthToRequest(deviceID int, authToRequest AccessDeviceAuthToRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceInfoRequest(deviceID int, infoRequest AccessDeviceInfoRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleSendAccessDeviceCommand(deviceID int, command ServerCommand) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleWelcomeRequest(makerspaceID int, deviceID int, cardTagID string) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) CommandAllAccessDevices(command ServerCommand) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) CommandMakerspaceAccessDevices(makerspaceID int, command ServerCommand) {
	panic("unimplemented")

}
