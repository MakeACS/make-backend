package auth

import (
	"context"
	"make-backend/internal/plugins/common"

	"github.com/hashicorp/go-plugin"
	grpc "google.golang.org/grpc"
)

type AuthProvider interface {
	common.BasePlugin
	Initialize(req PluginInitRequest)
	Heartbeat() common.HeartbeatInfo
	GetLoginURL(UserLoginStartRequest) LoginURL
	Logout(UserLogOffRequest)
}

type AuthPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	// Impl Injection
	Impl AuthProvider
}

func (p *AuthPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterAuthPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *AuthPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewAuthPluginClient(c)}, nil
}
