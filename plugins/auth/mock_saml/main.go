package main

/*
var pluginName = "auth.core.mock_saml"

const PluginName = "auth"

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   common.MagicKey,
	MagicCookieValue: "76d15ef6-1f0a-4e77-bff2-463daa54e19b",
}

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
