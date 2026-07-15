package acsmqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"make-backend/internal/api/acs"
	"make-backend/internal/database"
	"make-backend/internal/database/models"
	"make-backend/internal/logging"

	mqtt "github.com/mochi-mqtt/server/v2"
)

type MQTTController struct {
	ctx          context.Context
	slog         *slog.Logger
	logger       *logging.Logger
	store        *database.Store
	serverClient *mqtt.Server
}

func NewMQTTController(ctx context.Context, log *slog.Logger, store *database.Store, logger *logging.Logger, server *mqtt.Server) *MQTTController {
	controller := &MQTTController{
		ctx:          ctx,
		slog:         log,
		logger:       logger,
		store:        store,
		serverClient: server,
	}
	return controller
}

func (m *MQTTController) GetName() string {
	return "MQTTACSController"
}

func (m *MQTTController) sendResponseToTopic(topic string, data any) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal for send to topic '%s': %w", topic, err)
	}
	err = m.serverClient.Publish(topic, bs, true, 2)
	if err != nil {
		return fmt.Errorf("failed to send to topic '%s': %w", topic, err)
	}
	return nil
}

func (m *MQTTController) SendAccessDeviceAuthToResponse(access models.AccessDevice, response acs.ServerAuthToResponse) bool {
	topic := fmt.Sprintf("makerspace/device/%s/authTo/response", access.SN)
	err := m.sendResponseToTopic(topic, response)
	if err != nil {
		m.slog.Error("Failed to publish AccessDeviceAuthToResponse", "err", err)
		return false
	}
	return true
}

func (m *MQTTController) SendAccessDeviceCommand(dev models.AccessDevice, command acs.ServerCommand) bool {

	topic := fmt.Sprintf("makerspace/device/%s/authTo/response", dev.SN)
	err := m.sendResponseToTopic(topic, command)
	if err != nil {
		m.slog.Error("Failed to publish AccessDeviceCommand", "err", err)
		return false
	}
	return true
}

func (m *MQTTController) SendAccessDeviceConfigUpdate(dev models.AccessDevice, update acs.ServerConfigUpdateRequest) bool {
	topic := fmt.Sprintf("makerspace/device/%s/config/update", dev.SN)
	err := m.sendResponseToTopic(topic, update)
	if err != nil {
		m.slog.Error("Failed to publish AccessDeviceConfigUpdate", "err", err)
		return false
	}
	return true
}

func (m *MQTTController) SendAccessDeviceInfoResponse(dev models.AccessDevice, response acs.ServerInfoResponse) bool {
	topic := fmt.Sprintf("makerspace/device/%s/info/response", dev.SN)
	err := m.sendResponseToTopic(topic, response)
	if err != nil {
		m.slog.Error("Failed to publish AccessDeviceInfoResponse", "err", err)
		return false
	}
	return true
}

func (m *MQTTController) SendWelcomeResponse(dev models.Device, response acs.WelcomeResponse) bool {

	topic := fmt.Sprintf("makerspace/device/%s/welcome/response", dev.SN)
	err := m.sendResponseToTopic(topic, response)
	if err != nil {
		m.slog.Error("Failed to publish WelcomeResponse", "err", err)
		return false
	}
	return true

}
