package main

import (
	"fmt"
	"log"
	"make-backend/internal/plugins"
	"make-backend/internal/plugins/common"

	"github.com/hashicorp/go-plugin"
)

var pluginName = "notification.core.mock"

// Here is a real implementation of Greeter
type MockNotifier struct {
}

// NotifyUser implements [plugins.NotificationProvider].
func (m *MockNotifier) NotifyUser(args plugins.NotifyUserArgs) error {
	name := args.PreferredName
	email := args.Email
	s := fmt.Sprintf("To: %s\nDear %s,\n%s", email, name, args.Content.BodyHtml)
	log.Println("mock sending to single user: ", args.UserID, s)
	return nil

}

// Info implements [plugins.NotificationProvider].
func (m *MockNotifier) Info() plugins.PluginInfoResponse {
	// log.Println("info called")
	return plugins.PluginInfoResponse{
		Info: common.PluginInfo{
			Id:    pluginName,
			About: "plugin for not sending notifications but pretending to",
			Port:  0, // dont need any HTTP passthrough
		},
		Err: nil,
	}
}

// from example code:
//
//	handshakeConfigs are used to just do a basic handshake between
//	a plugin and host. If the handshake fails, a user friendly error is shown.
//	This prevents users from executing bad plugins or executing a plugin
//	directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   common.MagicKey,
	MagicCookieValue: "97bddc18-7cd3-4976-81cf-bcb5b756262a",
}

var _ plugins.NotificationProvider = &MockNotifier{}

func main() {
	notifier := &MockNotifier{}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		pluginName: &plugins.NotificationPlugin{Impl: notifier},
	}
	log.Println("Mock notification plugin started")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
