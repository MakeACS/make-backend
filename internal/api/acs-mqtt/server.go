package acsmqtt

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

var (
	serverUsername = os.Getenv("SERVER_MQTT_USERNAME")
	serverPassword = os.Getenv("SERVER_MQTT_PASSWORD")
)

func StartMqtt(port int) io.Closer {
	wsCfg := listeners.Config{
		Type:    listeners.TypeWS,
		ID:      "std-listener",
		Address: fmt.Sprintf(":%v", port)}

	server := mqtt.New(&mqtt.Options{
		Logger: slog.Default().With("server", "mqtt"),
	})

	if serverUsername == "" {
		log.Fatalf("Can not start MQTT Server without server username")
	}
	if serverPassword == "" {
		log.Fatalf("Can not start MQTT Server without server password")
	}

	_ = server.AddHook(new(auth.AllowHook), nil)

	wsListener := listeners.NewWebsocket(wsCfg)

	err := server.AddListener(wsListener)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	go server.Serve()
	return server
}
