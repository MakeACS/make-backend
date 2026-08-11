package main

import (
	"fmt"
	"log/slog"
	"make-backend/internal/plugins/auth"
	"make-backend/internal/plugins/common"
	"net"
	"net/http"

	"github.com/hashicorp/go-plugin"
)

var pluginName = "auth.core.saml"

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   common.MagicKey,
	MagicCookieValue: "76d15ef6-1f0a-4e77-bff2-463daa54e19b",
}

var Info = common.PluginInfo{
	Id:    pluginName,
	About: "A plugin for working with SAML SSO",
	Port:  0, // filled in later
}

type SAMLAuth struct{}

// GetLoginURL implements [auth.AuthProvider].
func (s *SAMLAuth) GetLoginURL(auth.UserLoginStartRequest) auth.LoginURL {
	panic("unimplemented")
}

// Heartbeat implements [auth.AuthProvider].
func (s *SAMLAuth) Heartbeat() common.HeartbeatInfo {
	panic("unimplemented")
}

// Initialize implements [auth.AuthProvider].
func (s *SAMLAuth) Initialize(req auth.PluginInitRequest) {
	panic("unimplemented")
}

// Logout implements [auth.AuthProvider].
func (s *SAMLAuth) Logout(auth.UserLogOffRequest) {
	panic("unimplemented")
}

func (s *SAMLAuth) Info() (*common.PluginInfo, error) {
	// log.Println("info called")
	return &Info, nil
}

func startHTTPHandle() (uint16, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
	// log.Println("SAML auth HTTP starting on port", port)

	go func() {
		if err := http.Serve(listener, nil); err != nil {
			slog.Error("server error", "err", err)
		}
	}()
	// log.Println("SAML auth HTTP started on port", port)

	return uint16(port), nil
}

func main() {
	// log.Println("SAML  starting")

	port, err := startHTTPHandle()
	if err != nil {
		// slog.Error("failed to start", "err", err)
		return
	}
	Info.Port = uint32(port)

	auth_s := &SAMLAuth{}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		pluginName: &auth.AuthPlugin{
			Impl: auth_s,
		},
	}
	// log.Println("SAML auth plugin started")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,

		GRPCServer: plugin.DefaultGRPCServer,
	})

}

var _ auth.AuthProvider = &SAMLAuth{}
