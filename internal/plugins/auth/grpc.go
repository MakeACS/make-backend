package auth

import (
	context "context"
	"make-backend/internal/plugins/common"
)

type GRPCClient struct{ client AuthPluginClient }

var _ AuthProvider = &GRPCClient{}

// Info implements [AuthProvider].
func (g *GRPCClient) Info() (*common.PluginInfo, error) {
	info, err := g.client.Info(context.Background(), nil)
	if err != nil {

		return nil, err
	}

	return info, nil

}

// GetLoginURL implements [AuthProvider].
func (g *GRPCClient) GetLoginURL(UserLoginStartRequest) LoginURL {
	panic("unimplemented")
}

// Heartbeat implements [AuthProvider].
func (g *GRPCClient) Heartbeat() common.HeartbeatInfo {
	panic("unimplemented")
}

// Initialize implements [AuthProvider].
func (g *GRPCClient) Initialize(req PluginInitRequest) {
	panic("unimplemented")
}

// Logout implements [AuthProvider].
func (g *GRPCClient) Logout(UserLogOffRequest) {
	panic("unimplemented")
}

// func (g *GRPCClient)

type GRPCServer struct {
	UnimplementedAuthPluginServer
	// This is the real implementation
	Impl AuthProvider
}

// GetLoginURL implements [AuthPluginServer].
func (g *GRPCServer) GetLoginURL(context.Context, *UserLoginStartRequest) (*LoginURL, error) {
	panic("unimplemented")
}

// Heartbeat implements [AuthPluginServer].
func (g *GRPCServer) Heartbeat(context.Context, *common.Empty) (*common.HeartbeatInfo, error) {
	panic("unimplemented")
}

// Info implements [AuthPluginServer].
func (g *GRPCServer) Info(ctx context.Context, _ *common.Empty) (*common.PluginInfo, error) {
	inf, err := g.Impl.Info()
	if err != nil {
		return nil, err
	}
	return inf, nil

}

// Initialize implements [AuthPluginServer].
func (g *GRPCServer) Initialize(context.Context, *PluginInitRequest) (*common.Empty, error) {
	panic("unimplemented")
}

// Logout implements [AuthPluginServer].
func (g *GRPCServer) Logout(context.Context, *UserLogOffRequest) (*common.Empty, error) {
	panic("unimplemented")
}

// mustEmbedUnimplementedAuthPluginServer implements [AuthPluginServer].
func (g *GRPCServer) mustEmbedUnimplementedAuthPluginServer() {
	panic("unimplemented")
}
