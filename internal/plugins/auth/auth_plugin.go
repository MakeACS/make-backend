package auth

import (
	"context"
	"make-backend/internal/plugins/common"

	"github.com/hashicorp/go-plugin"
	grpc "google.golang.org/grpc"
)

// What core exposes to the plugin
type AuthCallbackProvider interface {
	UserLoggedIn(*UserLoginCallback) (*RedirectURL, error)
	UserLoggedOut(*UserLogOffRequest) error
}

// What your plugin exposes
type AuthProvider interface {
	common.BasePlugin
	Heartbeat() common.HeartbeatInfo
	GetLoginURL(*UserLoginStartRequest) LoginURL
	Logout(*UserLogOffRequest)
	RegisterCallbackProvider(AuthCallbackProvider)
}

/*
 server start
         plugin start
 ask for addr
         give addr
 connect
 call initialize{channel}

 server -- init


*/

type AuthPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	// Impl Injection
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
