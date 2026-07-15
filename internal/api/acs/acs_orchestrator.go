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
	controllers map[int64]ACSController
	slog        *slog.Logger
	logger      *logging.Logger
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
		orc.slog.Warn("Couldn't find device for id for status report", "id", deviceID)
		return
	}
	dev.CurrentCardTag = statusReport.CurrentCardTag
	dev, err = orc.store.Devices.UpdateAccessDevice(context.TODO(), *dev)
	if err != nil {
		orc.slog.Warn("Failed to update card tag on status report for access device", "device_id", deviceID, "err", err)
	}
	for _, channel := range statusReport.Channels {
		_, err := orc.store.Devices.UpdateControllerState(context.TODO(), dev.Id, int(channel.ChannelID), channel.State)
		if err != nil {
			orc.slog.Warn("Failed to update channel state from status report", "device_id", deviceID, "channel_id", channel.ChannelID, "err", err)
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
		orc.slog.Warn("Couldn't find device for id for state change report", "id", deviceID)
		return
	}

	oldCardTag := dev.CurrentCardTag

	dev.CurrentCardTag = stateChangeReport.CurrentCardTag
	dev, err = orc.store.Devices.UpdateAccessDevice(context.TODO(), *dev)
	if err != nil {
		orc.slog.Warn("Failed to update card tag on state change report for access device", "devid", deviceID, "err", err)
	}

	for _, channel := range stateChangeReport.Channels {
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
				orc.slog.Warn("Failed to update channel state from state change report", "device_id", deviceID, "channel_id", channel.ChannelID, "err", err)
			}
		}

	}
}

func (orc *ACSOrchestrator) HandleAccessDeviceLogRequest(deviceID int64, logRequest AccessDeviceLogRequest) {
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
		// 	  await DeviceLogRepo.createDeviceLog(
		// 		core.deviceID,
		// 		DeviceLogSeverity.LOW,
		// 		{ type: "message", message: logRequest.message }
		// 	  )
	}

}

func HandleCoreConfigReport(deviceID int, configReport AccessDeviceConfigReport) {
	// try {
	//   const core = await CoreRepo.getCoreByDeviceID(deviceID);
	//   if (core === undefined) { return; }
	//   await core.updateConfiguration(configReport)
	// } catch (e) {
	//   await DeviceLogRepo.createDeviceLog(deviceID, DeviceLogSeverity.MEDIUM, { type: "core-config-report-error", error: e });
	// }
}

func (orc *ACSOrchestrator) HandleAccessDeviceAuthToRequest(deviceID int64, authToRequest AccessDeviceAuthToRequest) {
	panic("unimplemented")
	// try {
	// 	const core = await CoreRepo.getCoreByDeviceID(deviceID);
	// 	const user = await UserRepo.getUserByCardTagID(authToRequest.cardTagID);

	// 	if (core === undefined || user === undefined) {
	// 	  await AuditLogRepo.createAuditLog(
	// 		`{user} failed to activate {device}`,
	// 		"auth",
	// 		core?.makerspaceID,
	// 		{ id: user?.id ?? -1, label: user ? `${user.firstName} ${user.lastName}` : "Unknown User" },
	// 		{ id: deviceID, label: core ? core.name : "unkown device" }
	// 	  );

	// 	  if (core !== undefined) {
	// 		const controllers = await core.getAccessControllers();
	// 		ACSOrchestrator.getDeviceController(deviceID)?.sendCoreAuthToResponse(core, {
	// 		  channels: controllers.map((controller) => ({
	// 			channelID: controller.channelID,
	// 			state: authToRequest.state,
	// 			approved: false,
	// 			reason: AccessAttemptReason.UNKNOWN_USER
	// 		  })),
	// 		  cardTagID: authToRequest.cardTagID
	// 		})
	// 	  }
	// 	  return;
	// 	}

	// 	const attemptResult = await core.authTo(user.id, authToRequest.state, true);

	// 	ACSOrchestrator.getDeviceController(deviceID)?.sendCoreAuthToResponse(core, {
	// 	  channels: attemptResult,
	// 	  cardTagID: authToRequest.cardTagID
	// 	})
	//   } catch (e) {
	// 	await DeviceLogRepo.createDeviceLog(deviceID, DeviceLogSeverity.MEDIUM, { type: "core-authTo-error", error: e });
	//   }

}

func (orc *ACSOrchestrator) HandleAccessDeviceInfoRequest(deviceID int64, infoRequest AccessDeviceInfoRequest) {
	panic("unimplemented")
	// 	try {
	// 	  const core = await CoreRepo.getCoreByDeviceID(deviceID);
	// 	  if (core === undefined) { return; }

	// 	  const response: ServerInfoResponse = {
	// 		time: infoRequest.fields.includes(CoreInfoOptions.TIME)
	// 		  ? (new Date).getTime() : undefined,
	// 		state: infoRequest.fields.includes(CoreInfoOptions.STATE)
	// 		  ? (await core.getAccessControllers()).map((controller) => ({ id: controller.channelID, state: controller.state })) : undefined,
	// 		hmi: undefined,
	// 		flags: infoRequest.fields.includes(CoreInfoOptions.FLAGS)
	// 		  ? {
	// 			lockWhenIdle: core.flags?.lockWhenIdle ?? false,
	// 			restartWhenUnused: core.flags?.restartWhenUnused ?? false,
	// 			welcoming: (await core.getWelcomeMakerspace()) !== undefined
	// 		  } : undefined
	// 	  }

	// 	  ACSOrchestrator.getDeviceController(core.deviceID)?.sendCoreInfoResponse(core, response);
	// 	} catch (e) {
	// 	  await DeviceLogRepo.createDeviceLog(deviceID, DeviceLogSeverity.MEDIUM, { type: "core-info-request-error", error: e });
	// 	}

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
