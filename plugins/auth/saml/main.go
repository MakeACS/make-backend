package main

import (
	"bytes"
	"errors"
	"log/slog"
	"make-backend/internal/plugins/auth"
	"make-backend/internal/plugins/common"
	"net"
	"net/http"
	"net/url"

	"github.com/crewjam/saml/samlsp"
	"github.com/hashicorp/go-plugin"
)

var pluginName = "auth.core.saml"
var pluginAbout = "A plugin for working with SAML SSO"
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   common.MagicKey,
	MagicCookieValue: "76d15ef6-1f0a-4e77-bff2-463daa54e19b",
}

type SAMLAuth struct {
	saml         *samlsp.Middleware
	callbacks    auth.AuthCallbackProvider
	listener     net.Listener
	sp           SessionProviderViaPlugin
	status       common.PluginStatus
	statusString string
	config       Config
}

func (s *SAMLAuth) RegisterCallbackProvider(cb auth.AuthCallbackProvider) {
	s.sp = SessionProviderViaPlugin{
		cb:  cb,
		cfg: &s.config,
	}
	s.callbacks = cb

}

var _ http.ResponseWriter = &common.WriteAdapter{}

func (s *SAMLAuth) GetAuthDescription() (*auth.AuthDescription, error) {
	return &auth.AuthDescription{
		Name:     pluginName,
		Url:      s.config.BaseURL + "/login",
		ImageUrl: "https://www.shibboleth.net/wp-content/uploads/2020/10/shibboleth-icon-white-233x300.png",
	}, nil
}
func (s *SAMLAuth) GenerateLoginRequest(start *auth.UserLoginStartRequest) (*auth.LoginRequest, error) {
	if s.saml == nil {
		return nil, errors.New("SAML provider degraded")
	}
	w := common.WriteAdapter{
		Headers: http.Header{},
		Body:    bytes.Buffer{},
		Code:    0,
	}
	url, _ := url.Parse(start.GetOriginalURL())
	r := http.Request{
		URL: url,
	}
	s.saml.HandleStartAuthFlow(&w, &r)

	code, headers, body := w.IntoParts()

	req := auth.LoginRequest{}
	req.Code = int32(code)
	req.Body = body
	req.SetHeaders = []*auth.SetKV{}
	for k, vs := range headers {
		for _, v := range vs {
			req.SetHeaders = append(req.SetHeaders, &auth.SetKV{
				Key:   k,
				Value: v,
			})
		}
	}

	return &req, nil

}

func (s *SAMLAuth) Heartbeat() (*common.HeartbeatInfo, error) {
	hb := common.HeartbeatInfo{
		Status:        s.status,
		StatusMessage: s.statusString,
	}
	return &hb, nil
}

func (s *SAMLAuth) Logout(*auth.UserLogOffRequest) error {
	if s.saml == nil {
		return errors.New("SAML provider degraded")
	}
	// nothing to do on IDP side (for now)
	return nil
}

var (
	ConfigKeySpCert              = "SP_CERT"
	ConfigKeySpKey               = "SP_KEY"
	ConfigKeyIdpMetadataProvider = "IDP_METADATA_PROVIDER"
	ConfigKeyReplaceDomainFrom   = "SAML_REPLACE_DOMAIN_FROM"
	ConfigKeyReplaceDomainTo     = "SAML_REPLACE_DOMAIN_TO"
)

func SamlConfigFromPluginConfig(init *common.PluginInitialMessage) Config {
	var c Config
	c.BaseURL = init.PluginUrlBase
	for _, pair := range init.Configs {
		switch pair.Key {
		case ConfigKeySpCert:
			c.SPCert = pair.Value
		case ConfigKeySpKey:
			c.SPKey = pair.Value
		case ConfigKeyIdpMetadataProvider:
			c.SamlIDPMetadataProvider = pair.Value
		case ConfigKeyReplaceDomainFrom:
			c.OptionalReplaceDomainFrom = pair.Value
		case ConfigKeyReplaceDomainTo:
			c.OptionalReplaceDomainTo = pair.Value
		}
	}
	return c
}

func (s *SAMLAuth) Info(init *common.PluginInitialMessage) (*common.PluginInfo, error) {
	info := common.PluginInfo{
		Id:    pluginName,
		About: pluginAbout,
		Port:  0,
	}

	// only bring up if not up
	if s.saml == nil {
		cfg := SamlConfigFromPluginConfig(init)
		if cfg.SPCert == "" {
			s.status = common.PluginStatus_UNCONFIGURED
			s.statusString = "missing " + ConfigKeySpCert + " config key"
			return &info, errors.New(s.statusString)
		} else if cfg.SPKey == "" {
			s.status = common.PluginStatus_UNCONFIGURED
			s.statusString = "missing " + ConfigKeySpKey + " config key"
			return &info, errors.New(s.statusString)
		} else if cfg.SamlIDPMetadataProvider == "" {
			s.status = common.PluginStatus_UNCONFIGURED
			s.statusString = "missing " + ConfigKeyIdpMetadataProvider + " config key"
			return &info, errors.New(s.statusString)
		}
		var err error
		s.saml, err = s.SetupSamlSP(cfg, &s.sp)
		if err != nil {
			slog.Error("Failed to setup SAML SP connection", "err", err)
			s.status = common.PluginStatus_DEGRADED
			s.statusString = err.Error()
			return &info, err

		}
		s.startHTTPHandle()
	}
	port := s.listener.Addr().(*net.TCPAddr).Port
	info.Port = uint32(port)

	return &info, nil
}

func (s *SAMLAuth) startHTTPHandle() {
	http.DefaultServeMux.HandleFunc("/metadata", s.saml.ServeMetadata)
	http.DefaultServeMux.HandleFunc("/acs", s.saml.ServeACS)

	go func() {
		if err := http.Serve(s.listener, nil); err != nil {
			slog.Error("server error", "err", err)
			s.status = common.PluginStatus_DEGRADED
			s.statusString = "http server error: " + err.Error()
		}
	}()

}

func main() {
	// Claim a network port now so we know which to port to report to the core
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		slog.Error("failed to listen for http server", "err", err)
	}

	s_plugin := auth.AuthPlugin{
		Impl: &SAMLAuth{
			listener: listener,
		},
	}
	var pluginMap = map[string]plugin.Plugin{
		pluginName: &s_plugin,
	}
	// start serving plugin interface.
	// HTTP interface will be started when the core connects to us
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})

}

var _ auth.AuthProvider = &SAMLAuth{}
