package plugins

import (
	"fmt"
	"log/slog"
	"make-backend/internal/database"
	"make-backend/internal/plugins/auth"
	"make-backend/internal/plugins/common"
	"make-backend/internal/plugins/notify"
	"os"
	"os/exec"
	"path"

	"github.com/hashicorp/go-plugin"
)

var primaryNotificationProvider notify.NotificationProvider = nil

func PrimaryNotificationProvider() notify.NotificationProvider {
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
		Options:     map[string]string{},
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
	{
		Name:        "auth.core.saml",
		Version:     1,
		MagicValue:  "76d15ef6-1f0a-4e77-bff2-463daa54e19b",
		DownloadUrl: "",
		PluginType:  common.PluginType_Auth,
		Options: map[string]string{
			"SP_CERT":               os.Getenv("SAML_SP_CERT"),
			"SP_KEY":                os.Getenv("SAML_SP_KEY"),
			"IDP_METADATA_PROVIDER": os.Getenv("SAML_SP_METADATA_URL"),
		},
	},
}

func InterfaceForPluginType(t common.PluginType) plugin.Plugin {
	switch t {
	case common.PluginType_Notification:
		return &notify.NotificationPlugin{}
	case common.PluginType_Auth:
		return &auth.AuthPlugin{}
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

type TestAuthCBProvider struct{}

// UserLoggedIn implements [auth.AuthCallbackProvider].
func (t *TestAuthCBProvider) UserLoggedIn(*auth.UserLoginCallback) (*auth.RedirectURL, error) {
	slog.Info("User logged in")
	return nil, nil
}

// UserLoggedOut implements [auth.AuthCallbackProvider].
func (t *TestAuthCBProvider) UserLoggedOut(*auth.UserLogOffRequest) error {
	slog.Info("User logged out")
	return nil

}

var _ auth.AuthCallbackProvider = &TestAuthCBProvider{}

func initMessageForPlugin(host string, desc common.PluginDescription) *common.PluginInitialMessage {
	var msg = common.PluginInitialMessage{}
	msg.ServerHost = host
	msg.PluginUrlBase = fmt.Sprintf("%s/plugin/%s", host, desc.Name)

	msg.Configs = []*common.ConfigPair{}
	for k, v := range desc.Options {
		cfg := common.ConfigPair{
			Key:   k,
			Value: v,
		}
		msg.Configs = append(msg.Configs, &cfg)
	}
	return &msg
}

func StartPlugins(host string, store *database.Store) (func(), []PluginHTTPForwarding, error) {
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
			HandshakeConfig:  handshakeConfig,
			Plugins:          pluginMap,
			Cmd:              exec.Command(path.Join(plugin_dir, plugin_desc.Name)),
			Managed:          true, // Allow parent process (us) to kill clients when we leave
			SkipHostEnv:      true, // Dont leak secrets to plugins
			Logger:           common.NewPluginLogAdapter(*slog.Default().With("plugin", plugin_desc.Name), plugin_desc.Name, slog.LevelInfo),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
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

		msg := initMessageForPlugin(host, plugin_desc)
		base := raw.(common.BasePlugin)
		info, err := base.Info(msg)
		slog.Info("plugin start", "info", info)

		if err != nil {
			slog.Error("Failed to start plugin", "plugin", plugin_desc.Name, "err", err)
			continue
		}
		if info.Port != 0 {
			forwards = append(forwards, PluginHTTPForwarding{
				Path:   fmt.Sprintf("/plugin/%s", plugin_desc.Name),
				ToPort: uint16(info.Port),
			})
		}

		if plugin_desc.PluginType == common.PluginType_Auth {
			authPlugin, ok := raw.(auth.AuthProvider)
			if !ok {
				slog.Warn("plugin lied about type", "wanted", plugin_desc.PluginType, "plugin", plugin_desc.Name)
			}
			authPlugin.RegisterCallbackProvider(&TestAuthCBProvider{})
		}
	}

	return func() {
		plugin.CleanupClients()
	}, forwards, nil

}
