package plugins

import (
	"context"
	"fmt"
	"log/slog"
	"make-backend/internal/database"
	"os/exec"
	"path"

	"github.com/hashicorp/go-plugin"
)

var primaryNotificationProvider NotificationProvider = &nothingNotificationProvider{}

func PrimaryNotificationProvider() NotificationProvider {
	return primaryNotificationProvider
}

type PluginType string

var (
	PluginType_Auth         PluginType = "auth"
	PluginType_Currency     PluginType = "currency"
	PluginType_Notification PluginType = "notification"
)

type PluginInfo struct {
	Id    string
	About string
}

var MagicKey string = "ACS_PLUGIN_ID"

type PluginDescription struct {
	Name       string
	PluginType PluginType
	Version    uint
	MagicValue string
	Url        string /// empty url implies that the plugin will just be there and theres no need to download
}

// DB table somewhere
var wanted_plugins = []PluginDescription{
	{
		Name:       "notification.core.mock",
		Version:    1,
		MagicValue: "97bddc18-7cd3-4976-81cf-bcb5b756262a",
		Url:        "",
		PluginType: PluginType_Notification,
	},
	{
		Name:       "notification.core.linux_email",
		Version:    1,
		MagicValue: "904d74d0-24e7-49b8-a4f6-0914aa4edde8",
		Url:        "",
		PluginType: PluginType_Notification,
	},
}

func InterfaceForPluginType(t PluginType) plugin.Plugin {
	switch t {
	case PluginType_Notification:
		return &NotificationPlugin{}
	default:
		slog.Warn("unknown PluginType value", "type", t)
		return nil
	}
}

// generate the name:type mapping from the database plugin types
func generatePluginMap(wanted []PluginDescription) map[string]plugin.Plugin {
	pluginMap := map[string]plugin.Plugin{}
	for _, plugin_desc := range wanted {
		t := InterfaceForPluginType(plugin_desc.PluginType)
		if t != nil {
			pluginMap[plugin_desc.Name] = t
		}
	}
	return pluginMap
}

type userProviderFromStore struct {
	Store *database.Store
}

// EmailForUser implements [UserDataProvider].
func (u *userProviderFromStore) EmailForUser(userID int) (string, error) {
	user, err := u.Store.Users.GetUserById(context.TODO(), userID)
	if err != nil {
		return "", err
	}
	return user.Username + "@rit.edu", nil
}

// FullNameForUser implements [UserDataProvider].
func (u *userProviderFromStore) FullNameForUser(userID int) (string, error) {
	user, err := u.Store.Users.GetUserById(context.TODO(), userID)
	if err != nil {
		return "", err
	}
	return user.FullName(), nil

}

var _ UserDataProvider = &userProviderFromStore{}

func StartPlugins(store *database.Store) (func(), error) {
	plugin_dir := path.Join("./plugins", "bin")

	var pluginMap = generatePluginMap(wanted_plugins)

	for _, plugin_desc := range wanted_plugins {

		handshakeConfig := plugin.HandshakeConfig{
			ProtocolVersion:  plugin_desc.Version,
			MagicCookieKey:   MagicKey,
			MagicCookieValue: plugin_desc.MagicValue,
		}

		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path.Join(plugin_dir, plugin_desc.Name)),
			Managed:         true, // Allow parent process (us) to kill clients when we leave
			SkipHostEnv:     true, // Dont leak secrets to plugins
		})

		// Connect via RPC
		rpcClient, err := client.Client()
		if err != nil {
			slog.Warn("Failed to start plugin client", "plugin", plugin_desc.Name, "err", err)
			continue
		}

		// Request the plugin
		raw, err := rpcClient.Dispense(plugin_desc.Name)
		if err != nil {
			slog.Warn("Failed to dispense plugin", "plugin", plugin_desc.Name, "err", err)
			continue
		}

		notifier := raw.(NotificationProvider)
		infoErr := notifier.Info()
		fmt.Println("plugin info err", infoErr)

		err = notifier.NotifyUser(NotifyUserArgs{
			Provider: &userProviderFromStore{Store: store},
			UserID:   1,
			Content:  Notification{},
		})
		fmt.Println("sent message", err)

	}

	return func() {
		plugin.CleanupClients()
	}, nil

}
