package plugins

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type Notification struct {
	Title        string
	Body         string
	BodyHtml     string
	BodyMarkdown string
}

type PluginInfoResponse struct {
	Info PluginInfo
	Err  error
}

type NotifyUserArgs struct {
	Provider UserDataProvider
	UserID   int
	Content  Notification
}

type NotificationProvider interface {
	Info() PluginInfoResponse
	NotifyUser(NotifyUserArgs) error
	// NotifyUsersIndependently(provider UserDataProvider, userIds []int, notification Notification) error
	// NotifyGroup(provider UserDataProvider, groupId string, notification Notification) error
}

type nothingNotificationProvider struct{}

func (n *nothingNotificationProvider) NotifyUser(NotifyUserArgs) error {
	return errors.New("no notification plugin")
}

func (n *nothingNotificationProvider) Info() PluginInfoResponse {
	return PluginInfoResponse{Info: PluginInfo{}, Err: errors.New("no notification plugin")}
}

var _ NotificationProvider = &nothingNotificationProvider{}

type NotificationPlugin struct {
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

func (g *NotificationProviderRPC) Info() PluginInfoResponse {
	var info PluginInfoResponse
	err := g.client.Call("Plugin.Info", new(interface{}), &info)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		return PluginInfoResponse{info.Info, fmt.Errorf("failed to RPC call Info() : %w", err)}
	}

	return info
}

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type NotificationProviderRPCServer struct {
	// This is the real implementation
	Impl NotificationProvider
}

func (s *NotificationProviderRPCServer) Info(args interface{}, resp *PluginInfoResponse) error {
	*resp = s.Impl.Info()
	return nil
}

func (p *NotificationPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &NotificationProviderRPCServer{Impl: p.Impl}, nil
}

func (NotificationPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &NotificationProviderRPC{client: c}, nil
}
