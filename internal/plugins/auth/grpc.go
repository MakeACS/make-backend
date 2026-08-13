package auth

import (
	context "context"
	"fmt"
	"log/slog"
	"make-backend/internal/plugins/common"

	"github.com/hashicorp/go-plugin"
	grpc "google.golang.org/grpc"
)

type GRPCClient struct {
	client AuthPluginClient

	broker         *plugin.GRPCBroker
	callbackServer AuthCallbackServiceServer
}

// RegisterCallbackProvider implements [AuthProvider].
func (g *GRPCClient) RegisterCallbackProvider(cb AuthCallbackProvider) {
	g.callbackServer = &GRPCCallbackServer{Impl: cb}
	err := g.initialize()
	if err != nil {
		slog.Warn("could not init auth plugin", "err", err)
	}

}

type GRPCCallbackClient struct {
	client AuthCallbackServiceClient
}
type GRPCCallbackServer struct {
	UnimplementedAuthCallbackServiceServer
	// This is the real implementation
	Impl AuthCallbackProvider
}

func (g *GRPCCallbackServer) UserLoggedIn(c context.Context, cb *UserLoginCallback) (*RedirectURL, error) {
	return g.Impl.UserLoggedIn(cb)
}
func (g *GRPCCallbackServer) UserLoggedOut(c context.Context, req *UserLogOffRequest) (*common.Empty, error) {
	err := g.Impl.UserLoggedOut(req)
	empty := common.Empty{}
	return &empty, err
}

// UserLoggedIn implements [AuthCallbackProvider].
func (g *GRPCCallbackClient) UserLoggedIn(cb *UserLoginCallback) (*RedirectURL, error) {
	u, err := g.client.UserLoggedIn(context.TODO(), &UserLoginCallback{
		Email:             cb.Email,
		FullName:          cb.FullName,
		PreferredName:     cb.PreferredName,
		ProfilePictureUrl: cb.ProfilePictureUrl,
		OriginalURL:       cb.OriginalURL,
	})
	return u, err

}

// UserLoggedOut implements [AuthCallbackProvider].
func (g *GRPCCallbackClient) UserLoggedOut(*UserLogOffRequest) error {
	panic("unimplemented")
}

var _ AuthCallbackProvider = &GRPCCallbackClient{}
var _ AuthProvider = &GRPCClient{}

func (g *GRPCClient) initialize() error {

	callbackBrokerID := g.broker.NextId()

	go g.broker.AcceptAndServe(callbackBrokerID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		RegisterAuthCallbackServiceServer(server, g.callbackServer)
		return server
	})

	_, err := g.client.InternalInitializeCallbacks(context.TODO(), &PluginInitRequest{
		CallbackBrokerId: uint64(callbackBrokerID),
	})
	return err
}

// Info implements [AuthProvider].
func (g *GRPCClient) Info(init *common.PluginInitialMessage) (*common.PluginInfo, error) {
	info, err := g.client.Info(context.Background(), init)
	if err != nil {

		return nil, err
	}

	return info, nil

}

// GenerateLoginRequest implements [AuthProvider].
func (g *GRPCClient) GenerateLoginRequest(req *UserLoginStartRequest) (*LoginRequest, error) {
	res, err := g.client.GetLoginRequest(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return res, nil

}

// Heartbeat implements [AuthProvider].
func (g *GRPCClient) Heartbeat() common.HeartbeatInfo {
	panic("unimplemented")
}

// Logout implements [AuthProvider].
func (g *GRPCClient) Logout(*UserLogOffRequest) {
	panic("unimplemented")
}

type GRPCServer struct {
	UnimplementedAuthPluginServer
	// This is the real implementation
	Impl   AuthProvider
	broker *plugin.GRPCBroker
}

func (g *GRPCServer) GetLoginRequest(ctx context.Context, req *UserLoginStartRequest) (*LoginRequest, error) {
	res, err := g.Impl.GenerateLoginRequest(req)
	if err != nil {
		return nil, err
	}
	return res, nil

}

// Heartbeat implements [AuthPluginServer].
func (g *GRPCServer) Heartbeat(context.Context, *common.Empty) (*common.HeartbeatInfo, error) {
	panic("unimplemented")
}

// Info implements [AuthPluginServer].
func (g *GRPCServer) Info(ctx context.Context, init *common.PluginInitialMessage) (*common.PluginInfo, error) {
	inf, err := g.Impl.Info(init)
	if err != nil {
		return nil, err
	}
	return inf, nil

}

func (g *GRPCServer) InternalInitializeCallbacks(ctx context.Context, req *PluginInitRequest) (*common.Empty, error) {
	slog.Debug("Auth GRPC server init callbacks", "broker_id", req.GetCallbackBrokerId())
	conn, err := g.broker.Dial(uint32(req.GetCallbackBrokerId()))
	if err != nil {
		return nil, fmt.Errorf("dial host callback broker: %w", err)
	}

	cbClient := GRPCCallbackClient{
		client: NewAuthCallbackServiceClient(conn),
	}
	g.Impl.RegisterCallbackProvider(&cbClient)

	empty := common.Empty{}
	return &empty, nil
}

// Logout implements [AuthPluginServer].
func (g *GRPCServer) Logout(context.Context, *UserLogOffRequest) (*common.Empty, error) {
	panic("unimplemented")
}
