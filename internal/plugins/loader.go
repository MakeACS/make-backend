package plugins

import (
	"fmt"
	"log/slog"
	"make-backend/internal/database"
	"make-backend/internal/plugins/common"
	"os/exec"
	"path"

	"github.com/hashicorp/go-plugin"
)

var primaryNotificationProvider NotificationProvider = &nothingNotificationProvider{}

func PrimaryNotificationProvider() NotificationProvider {
	return primaryNotificationProvider
}

// DB table somewhere
var wanted_plugins = []common.PluginDescription{
	{
		Name:        "notification.core.mock",
		Version:     1,
		MagicValue:  "97bddc18-7cd3-4976-81cf-bcb5b756262a",
		DownloadUrl: "",
		PluginType:  common.PluginType_Notification,
	},
	// {
	// 	Name:       "notification.core.linux_email",
	// 	Version:    1,
	// 	MagicValue: "904d74d0-24e7-49b8-a4f6-0914aa4edde8",
	// 	Url:        "",
	// 	PluginType: PluginType_Notification,
	// },
	// {
	// 	Name:       "auth.rit.shibboleth_sso",
	// 	Version:    1,
	// 	MagicValue: "904d74d0-24e7-49b8-a4f6-0914aa4edde8",
	// 	Url:        "",
	// 	PluginType: PluginType_Notification,
	// },
	// {
	// 	Name:        "notification.base.mock_saml",
	// 	Version:     1,
	// 	MagicValue:  "76d15ef6-1f0a-4e77-bff2-463daa54e19b",
	// 	DownloadUrl: "",
	// 	PluginType:  common.PluginType_Auth,
	// },
}

func InterfaceForPluginType(t common.PluginType) plugin.Plugin {
	switch t {
	case common.PluginType_Notification:
		return &NotificationPlugin{}
	// case PluginType_Auth:
	// return &AuthPlugin{}
	default:
		slog.Warn("unknown PluginType value", "type", t)
		return nil
	}
}

// generate the name:type mapping from the database plugin types
func generatePluginMap(wanted []common.PluginDescription) map[string]plugin.Plugin {
	pluginMap := map[string]plugin.Plugin{}
	for _, plugin_desc := range wanted {
		t := InterfaceForPluginType(plugin_desc.PluginType)
		if t != nil {
			pluginMap[plugin_desc.Name] = t
		}
	}
	return pluginMap
}

func StartPlugins(store *database.Store) (func(), []PluginHTTPForwarding, error) {
	plugin_dir := path.Join("./plugins", "bin")

	forwards := []PluginHTTPForwarding{}

	var pluginMap = generatePluginMap(wanted_plugins)

	for _, plugin_desc := range wanted_plugins {

		handshakeConfig := plugin.HandshakeConfig{
			ProtocolVersion:  plugin_desc.Version,
			MagicCookieKey:   common.MagicKey,
			MagicCookieValue: plugin_desc.MagicValue,
		}

		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path.Join(plugin_dir, plugin_desc.Name)),
			Managed:         true, // Allow parent process (us) to kill clients when we leave
			SkipHostEnv:     true, // Dont leak secrets to plugins
			Logger:          common.NewPluginLogAdapter(*slog.Default(), plugin_desc.Name, slog.LevelInfo),
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

		notifier := raw.(BasePlugin)
		infoErr := notifier.Info()

		if infoErr.Err != nil {
			slog.Error("Failed to start plugin", "plugin", plugin_desc.Name, "err", infoErr.Err)
			continue
		}
		if infoErr.Info.Port != 0 {
			forwards = append(forwards, PluginHTTPForwarding{
				Path:   fmt.Sprintf("/plugin/%s", plugin_desc.Name),
				ToPort: infoErr.Info.Port,
			})
		}
	}

	return func() {
		plugin.CleanupClients()
	}, forwards, nil

}
