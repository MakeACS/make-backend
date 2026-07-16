package acsmqtt

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"make-backend/internal/api/acs"
	"make-backend/internal/database"
	"make-backend/internal/logging"
	"os"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

var (
	maybeMqttAdminUsername = os.Getenv("MQTT_ADMIN_USERNAME")
	maybeMqttAdminPassword = os.Getenv("MQTT_ADMIN_PASSWORD")
	runtimeContext         = os.Getenv("RUNTIME_CONTEXT") // development/staging/production
)

func mqttAdminAccountEnabled() bool {
	return maybeMqttAdminUsername != "" && maybeMqttAdminPassword != ""
}

func StartMqtt(logger *logging.Logger, store *database.Store, port int) (io.Closer, acs.ACSController) {
	wsCfg := listeners.Config{
		Type:    listeners.TypeWS,
		ID:      "std-listener",
		Address: fmt.Sprintf(":%v", port)}

	server := mqtt.New(&mqtt.Options{
		Logger:       slog.Default().With("server", "mqtt"),
		InlineClient: true,
	})

	if !mqttAdminAccountEnabled() {
		log.Printf("Not enabling admin user for MQTT bc no username or password")
	} else {
		log.Printf("Enabling MQTT admin access")
	}
	if runtimeContext == "" {
		log.Fatalf("Can not start MQTT Server without RUNTIME_CONTEXT to determine write access")
	}

	_ = server.AddHook(&AuthHook{store: store}, nil)

	wsListener := listeners.NewWebsocket(wsCfg)

	err := server.AddListener(wsListener)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	go server.Serve()
	return server, NewMQTTController(context.TODO(), slog.Default(), store, logger, server)
}
