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

/*
// CallbackServer is implemented by the host application. The plugin gets a
// brokered gRPC client for this interface after Initialize is called.
type CallbackServer interface {
	auth.AuthCallbackServiceServer
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl           auth.AuthPluginServer
	CallbackServer CallbackServer
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	auth.RegisterAuthPluginServer(server, p.Impl)
	return nil
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *auth.ClientConn) (interface{}, error) {
	return &Client{
		client:         auth.NewAuthPluginClient(conn),
		broker:         broker,
		callbackServer: p.CallbackServer,
	}, nil
}

var _ plugin.GRPCPlugin = (*Plugin)(nil)

// Client is the host-side wrapper. It owns the brokered callback server and
// exposes the generated AuthPlugin client.
type Client struct {
	client           auth.AuthPluginClient
	broker           *plugin.GRPCBroker
	callbackServer   CallbackServer
	callbackBrokerID uint32
}

func (c *Client) Initialize(ctx context.Context) error {
	if c.callbackServer == nil {
		return fmt.Errorf("callback server is nil")
	}

	c.callbackBrokerID = c.broker.NextId()

	go c.broker.AcceptAndServe(c.callbackBrokerID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		auth.RegisterAuthCallbackServiceServer(server, c.callbackServer)
		return server
	})

	// // _, err := c.client.Initialize(ctx, &auth.PluginInitRequest{
	// // 	CallbackBrokerId: uint64(c.callbackBrokerID),
	// // })
	// return err
	return nil
}

func (c *Client) GetLoginURL(ctx context.Context, req *auth.UserLoginStartRequest) (*auth.LoginURL, error) {
	return c.client.GetLoginURL(ctx, req)
}

func (c *Client) Logout(ctx context.Context, req *auth.UserLogOffRequest) error {
	_, err := c.client.Logout(ctx, req)
	return err
}

func main() {
	// impl, err := authplugin.NewSAMLAuthPlugin()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// p := &authplugin.Plugin{Impl: impl}

	// The broker is available when GRPCServer is invoked.
	// Wrap the plugin so it can inject it into the implementation.
	pImpl := &serverPlugin{}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         map[string]plugin.GRPCPlugin{PluginName: pImpl},
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type serverPlugin struct {
	// *authplugin.Plugin
	// impl *authplugin.SAMLAuthPlugin
}

func (p *serverPlugin) GRPCServer(b *plugin.GRPCBroker, s *grpc.Server) error {
	p.impl.SetBroker(b)
	return p.Plugin.GRPCServer(b, s)
}
*/
