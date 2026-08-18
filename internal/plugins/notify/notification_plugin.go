package notify

import (
	context "context"
	"fmt"
	"make-backend/internal/plugins/common"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	grpc "google.golang.org/grpc"
)

type Notification struct {
	Title        string
	Body         string
	BodyHtml     string
	BodyMarkdown string
}

type NotifyUserArgs struct {
	UserID        int
	Email         string
	PreferredName string
	Content       Notification
}

type NotificationProvider interface {
	common.BasePlugin
	NotifyUser(NotifyUserArgs) error
}

type NotificationPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	// Impl Injection
	Impl NotificationProvider
}

// Here is an implementation that talks over RPC
type NotificationProviderRPC struct{ client *rpc.Client }

// NotifyUser implements [NotificationProvider].
func (g *NotificationProviderRPC) NotifyUser(arg NotifyUserArgs) error {
	var farErr error
	err := g.client.Call("Plugin.NotifyUser", &arg, &farErr)
	if err != nil {
		return fmt.Errorf("failed to RPC call NotifyUser: %w", err)
	}
	return farErr
}

var _ NotificationProvider = &NotificationProviderRPC{}

func (g *NotificationProviderRPC) Info(*common.PluginInitialMessage) (*common.PluginInfo, error) {
	panic("broken with grpc migration")
	// var info common.PluginInfo
	// err := g.client.Call("Plugin.Info", new(interface{}), &info)
	// if err != nil {
	// 	// You usually want your interfaces to return errors. If they don't,
	// 	// there isn't much other choice here.
	// 	return common.PluginInfoResponse{info.Info, fmt.Errorf("failed to RPC call Info() : %w", err)}
	// }

	// return info
}

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type NotificationProviderRPCServer struct {
	// This is the real implementation
	Impl NotificationProvider
}

func (s *NotificationProviderRPCServer) NotifyUser(args NotifyUserArgs, resp *error) error {
	arg := NotifyUserArgs{
		UserID:        2,
		Email:         "test2",
		PreferredName: "test 2",
		Content:       Notification{},
	}
	return s.Impl.NotifyUser(arg)
}

func (s *NotificationProviderRPCServer) Info(args *common.PluginInitialMessage, resp *common.PluginInfo) error {
	r, err := s.Impl.Info(args)
	resp.Id = r.Id
	resp.About = r.About
	resp.Port = r.Port
	return err
}
func (p *NotificationPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterNotificationPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *NotificationPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewNotificationPluginClient(c)}, nil
}
