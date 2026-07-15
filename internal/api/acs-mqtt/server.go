package acsmqtt

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"make-backend/internal/logging"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func StartMqtt(logger *logging.Logger, port int) io.Closer {
	wsCfg := listeners.Config{
		Type:    listeners.TypeWS,
		ID:      "std-listener",
		Address: fmt.Sprintf(":%v", port)}

	server := mqtt.New(&mqtt.Options{
		Logger: slog.Default().With("server", "mqtt"),
	})

	_ = server.AddHook(new(auth.AllowHook), nil)

	wsListener := listeners.NewWebsocket(wsCfg)

	err := server.AddListener(wsListener)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	go server.Serve()
	return server
}
