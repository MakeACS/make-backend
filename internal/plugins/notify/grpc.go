package notify

import (
	context "context"
	"make-backend/internal/plugins/common"
)

type GRPCClient struct{ client NotificationPluginClient }

// NotifyUser implements [NotificationProvider].
func (g *GRPCClient) NotifyUser(NotifyUserArgs) error {
	panic("unimplemented")
}

var _ NotificationProvider = &GRPCClient{}

// Info implements [AuthProvider].
func (g *GRPCClient) Info() (*common.PluginInfo, error) {
	info, err := g.client.Info(context.Background(), nil)
	if err != nil {

		return nil, err
	}

	return info, nil

}

type GRPCServer struct {
	UnimplementedNotificationPluginServer
	// This is the real implementation
	Impl NotificationProvider
}

// Info implements [AuthPluginServer].
func (g *GRPCServer) Info(ctx context.Context, _ *common.Empty) (*common.PluginInfo, error) {
	inf, err := g.Impl.Info()
	if err != nil {
		return nil, err
	}
	return inf, nil

}
