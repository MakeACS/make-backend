package main

import (
	"fmt"
	"log"
	"make-backend/internal/plugins"
	"make-backend/internal/plugins/common"

	"github.com/hashicorp/go-plugin"
)

var PluginId = "notification.core.linux_email"

type EmailNotifier struct {
}

// NotifyUser implements [plugins.NotificationProvider].
func (m *EmailNotifier) NotifyUser(args plugins.NotifyUserArgs) error {
	email := args.Email
	name := args.PreferredName
	s := fmt.Sprintf("To: %s\nDear %s,\n%s", email, name, args.Content.BodyHtml)
	log.Println("Sending email to single user: ", args.UserID, s)
	return nil

}

// Info implements [plugins.NotificationProvider].
func (m *EmailNotifier) Info() plugins.PluginInfoResponse {
	// log.Println("info called")
	return plugins.PluginInfoResponse{
		Info: common.PluginInfo{
			Id:    PluginId,
			About: "plugin for sending notifications via email",
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
	MagicCookieValue: "904d74d0-24e7-49b8-a4f6-0914aa4edde8",
}

var _ plugins.NotificationProvider = &EmailNotifier{}

func main() {
	notifier := &EmailNotifier{}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		PluginId: &plugins.NotificationPlugin{Impl: notifier},
	}
	log.Println("Email notification plugin started")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
