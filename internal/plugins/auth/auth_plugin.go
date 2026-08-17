package auth

import (
	"context"
	"make-backend/internal/plugins/common"

	"github.com/hashicorp/go-plugin"
	grpc "google.golang.org/grpc"
)

// What core exposes to the plugin
type AuthCallbackProvider interface {
	// When a auth plugin confirms a user's credentials, it calls this to start a new session in the core and find where it should redirect the user
	UserLoggedIn(*UserLoginCallback) (*RedirectURL, error)
	// Possibly not needed/TODO
	// When a user finishes logging out of the auth system, call this to delete the session in the core
	UserLoggedOut(*UserLogOffRequest) error
}

// What auth plugins expose/what the core can ask
type AuthProvider interface {
	common.BasePlugin
	// General status information about the plugin and if its still working
	Heartbeat() (*common.HeartbeatInfo, error)
	// Generate a URL/Body/headers for
	GenerateLoginRequest(*UserLoginStartRequest) (*LoginRequest, error)
	// Called when a user logs out to end any sessions in the auth provider
	Logout(*UserLogOffRequest) error
	// Pass a provider from plugin user (core) to the plugin so it can talk back
	RegisterCallbackProvider(AuthCallbackProvider)
}

type AuthPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl AuthProvider
}

func (p *AuthPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterAuthPluginServer(s, &GRPCServer{
		Impl:   p.Impl,
		broker: broker,
	})
	return nil
}

func (p *AuthPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client:         NewAuthPluginClient(c),
		broker:         broker,
		callbackServer: nil,
	}, nil
}
