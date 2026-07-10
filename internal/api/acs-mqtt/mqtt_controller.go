package acsmqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"make-backend/internal/api/acs"
	"make-backend/internal/database"
	"make-backend/internal/database/models"

	mqtt "github.com/mochi-mqtt/server/v2"
)

type MQTTController struct {
	ctx          context.Context
	log          *slog.Logger
	store        *database.Store
	serverClient *mqtt.Server
}

func NewMQTTController(ctx context.Context, log *slog.Logger, store *database.Store, server *mqtt.Server) *MQTTController {
	return &MQTTController{
		ctx:          ctx,
		log:          log,
		store:        store,
		serverClient: server,
	}
}

func (m *MQTTController) GetName() string {
	return "MQTTACSController"
}

func (m *MQTTController) SendAccessDeviceAuthToResponse(access models.AccessDevice, response acs.ServerAuthToResponse) bool {

	topic := fmt.Sprintf("makerspace/device/%s/authTo/response", access.SN)
	bs, err := json.Marshal(response)
	if err != nil {
		m.log.Error("Failed to marshal AccessDeviceAuthToResponse", "err", err)
		return false
	}
	err = m.serverClient.Publish(topic, bs, true, 2)
	if err != nil {
		m.log.Error("Failed to publish AccessDeviceAuthToResponse", "err", err)
		return false
	}
	return true
}

func (m *MQTTController) SendAccessDeviceCommand(adev models.AccessDevice, command acs.ServerCommand) bool {
	panic("unimplemented")
}

func (m *MQTTController) SendAccessDeviceConfigUpdate(adev models.AccessDevice, update acs.ServerConfigUpdateRequest) bool {
	panic("unimplemented")
}

func (m *MQTTController) SendAccessDeviceInfoResponse(adev models.AccessDevice, response acs.ServerInfoResponse) bool {
	panic("unimplemented")
}

func (m *MQTTController) SendWelcomeResponse(device models.Device, response acs.WelcomeResponse) bool {
	panic("unimplemented")
}
