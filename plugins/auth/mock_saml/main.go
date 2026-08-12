package main

import (
	"fmt"
	"log"
	"log/slog"
	"make-backend/internal/plugins/auth"
	"make-backend/internal/plugins/common"
	"net"
	"net/http"

	"github.com/crewjam/saml/samlsp"
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

type SAMLAuth struct {
	callbacks auth.AuthCallbackProvider
	listener  net.Listener
	// pluginstate
}

// RegisterCallbackProvider implements [auth.AuthProvider].
func (s *SAMLAuth) RegisterCallbackProvider(cb auth.AuthCallbackProvider) {
	s.callbacks = cb
}

// GetLoginURL implements [auth.AuthProvider].
func (s *SAMLAuth) GetLoginURL(*auth.UserLoginStartRequest) (*auth.LoginURL, error) {
	return &auth.LoginURL{Url: "http://http.cat/404"}, nil

}

// Heartbeat implements [auth.AuthProvider].
func (s *SAMLAuth) Heartbeat() common.HeartbeatInfo {
	panic("unimplemented")
}

// Initialize implements [auth.AuthProvider].
func (s *SAMLAuth) Initialize(req *auth.PluginInitRequest) error {
	return fmt.Errorf("initialize called when it was supposed to be intercepted")
}

// Logout implements [auth.AuthProvider].
func (s *SAMLAuth) Logout(*auth.UserLogOffRequest) {
	panic("unimplemented")
}

func SamlConfigFromPluginConfig(init *common.PluginInitialMessage) Config {
	var c Config
	c.BaseURL = init.PluginUrlBase
	log.Println("CHost", c.BaseURL)
	for _, pair := range init.Configs {
		switch pair.Key {
		case "SP_CERT":
			c.SPCert = pair.Value
		case "SP_KEY":
			c.SPKey = pair.Value
		case "IDP_METADATA_PROVIDER":
			c.SamlIDPMetadataProvider = pair.Value
		}
	}
	return c
}

func (s *SAMLAuth) Info(init *common.PluginInitialMessage) (*common.PluginInfo, error) {
	endpoints := SetupSamlSP(SamlConfigFromPluginConfig(init))
	err := startHTTPHandle(s.listener, endpoints)
	if err != nil {
		// slog.Error("failed to start", "err", err)
		return nil, err
	}
	port := s.listener.Addr().(*net.TCPAddr).Port

	Info.Port = uint32(port)
	return &Info, nil
}

func startHTTPHandle(listener net.Listener, handler *samlsp.Middleware) error {

	http.DefaultServeMux.HandleFunc("/metadata", handler.ServeMetadata)
	http.DefaultServeMux.HandleFunc("/acs", handler.ServeACS)
	// http.Handle("/plugin/auth.core.saml/", handler)
	http.Handle("/protected", handler.RequireAccount(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("protected"))
	})))
	go func() {
		if err := http.Serve(listener, nil); err != nil {
			slog.Error("server error", "err", err)
		}
	}()

	return nil
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	auth_s := &SAMLAuth{
		listener: listener,
	}
	s_plugin := auth.AuthPlugin{
		Impl: auth_s,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		pluginName: &s_plugin,
	}
	// log.Println("SAML auth plugin started")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,

		GRPCServer: plugin.DefaultGRPCServer,
	})

}

var _ auth.AuthProvider = &SAMLAuth{}
