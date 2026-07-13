package acs

import (
	"context"
	"log/slog"
	"make-backend/internal/database"
	"make-backend/internal/database/models"
)

type ACSOrchestrator struct {
	store       *database.Store
	controllers map[int64]ACSController
	log         *slog.Logger
}

func (orc *ACSOrchestrator) RegisterDevice(deviceID int64, controller ACSController) {
	orc.controllers[deviceID] = controller
}

func (orc *ACSOrchestrator) GetDeviceController(deviceID int64) ACSController {
	val, ok := orc.controllers[deviceID]
	if !ok {
		return nil
	}
	return val
}

func (orc *ACSOrchestrator) HandleAccessDeviceStatusReport(deviceID int64, statusReport AccessDeviceStatusReport) {
	dev, err := orc.store.Devices.GetAccessDeviceById(context.TODO(), int(deviceID))
	if err != nil {
		orc.log.Warn("Couldn't find device for id for status report", "id", deviceID)
		return
	}
	dev.CurrentCardTag = statusReport.CurrentCardTag
	dev, err = orc.store.Devices.UpdateAccessDevice(context.TODO(), *dev)
	if err != nil {
		orc.log.Warn("Failed to update card tag on status report for access device", "devid", deviceID, "err", err)
	}
	for _, channel := range statusReport.Channels {
		_, err := orc.store.Devices.UpdateControllerState(context.TODO(), dev.Id, int(channel.ChannelID), channel.State)
		if err != nil {
			orc.log.Warn("Failed to update channel state from status report", "devid", deviceID, "channelid", channel.ChannelID, "err", err)
		}
	}
}

func (orc *ACSOrchestrator) endSession(dev *models.AccessDevice, oldCardTag string) {
	panic("unimplemented")
}
func (orc *ACSOrchestrator) startSession(dev *models.AccessDevice, cardTag string) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceStateChangeReport(deviceID int64, stateChangeReport AccessDeviceStateChangeReport) {
	dev, err := orc.store.Devices.GetAccessDeviceById(context.TODO(), int(deviceID))
	if err != nil {
		orc.log.Warn("Couldn't find device for id for state change report", "id", deviceID)
		return
	}

	oldCardTag := dev.CurrentCardTag

	dev.CurrentCardTag = stateChangeReport.CurrentCardTag
	dev, err = orc.store.Devices.UpdateAccessDevice(context.TODO(), *dev)
	if err != nil {
		orc.log.Warn("Failed to update card tag on state change report for access device", "devid", deviceID, "err", err)
	}

	for i, channel := range stateChangeReport.Channels {
		if channel.FromState == models.StateUnlocked {
			// (await ACRepo.getAccessControllersByDeviceAndChannelID(deviceID, channel.channelID))?.endSession(oldCardTag ?? "");
			orc.endSession(dev, oldCardTag)
		} else if channel.ToState == models.StateUnlocked {
			// startSession
			orc.startSession(dev, stateChangeReport.CurrentCardTag)
		}
		for _, channel := range stateChangeReport.Channels {
			_, err := orc.store.Devices.UpdateControllerState(context.TODO(), dev.Id, int(channel.ChannelID), channel.ToState)
			if err != nil {
				orc.log.Warn("Failed to update channel state from state change report", "devid", deviceID, "channelid", channel.ChannelID, "err", err)
			}
		}

	}
}

func (orc *ACSOrchestrator) HandleAccessDeviceLogRequest(deviceID int64, logRequest AccessDeviceLogRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceAuthToRequest(deviceID int64, authToRequest AccessDeviceAuthToRequest) {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleAccessDeviceInfoRequest(deviceID int64, infoRequest AccessDeviceInfoRequest) {
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
