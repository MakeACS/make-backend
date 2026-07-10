package acsmqtt

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"make-backend/internal/api/acs"
	"make-backend/internal/database"
	"os"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

var (
	serverUsername = os.Getenv("SERVER_MQTT_USERNAME")
	serverPassword = os.Getenv("SERVER_MQTT_PASSWORD")
)

func StartMqtt(port int, store *database.Store) (io.Closer, acs.ACSController) {
	wsCfg := listeners.Config{
		Type:    listeners.TypeWS,
		ID:      "std-listener",
		Address: fmt.Sprintf(":%v", port)}

	server := mqtt.New(&mqtt.Options{
		Logger:       slog.Default().With("server", "mqtt"),
		InlineClient: true,
	})

	if serverUsername == "" {
		log.Fatalf("Can not start MQTT Server without server username")
	}
	if serverPassword == "" {
		log.Fatalf("Can not start MQTT Server without server password")
	}

	_ = server.AddHook(&AuthHook{store: store}, nil)

	wsListener := listeners.NewWebsocket(wsCfg)

	err := server.AddListener(wsListener)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	go server.Serve()
	return server, NewMQTTController(context.TODO(), slog.Default(), store, server)
}
