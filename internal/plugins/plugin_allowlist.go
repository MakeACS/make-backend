package plugins

import (
	"log/slog"
	"make-backend/internal/plugins/auth"
	"make-backend/internal/plugins/common"
	"make-backend/internal/plugins/notify"
	"os"

	"github.com/hashicorp/go-plugin"
)

// eventually a DB table and loaded from the store
var wanted_plugins = []common.PluginDescription{
	{
		Name:        "notification.core.mock",
		Version:     1,
		MagicValue:  "97bddc18-7cd3-4976-81cf-bcb5b756262a",
		DownloadUrl: "",
		PluginType:  common.PluginType_Notification,
		Options:     map[string]string{},
	},
	{
		Name:        "auth.core.saml",
		Version:     1,
		MagicValue:  "76d15ef6-1f0a-4e77-bff2-463daa54e19b",
		DownloadUrl: "",
		PluginType:  common.PluginType_Auth,
		Options: map[string]string{
			"SP_CERT":                  os.Getenv("SAML_SP_CERT"),
			"SP_KEY":                   os.Getenv("SAML_SP_KEY"),
			"IDP_METADATA_PROVIDER":    os.Getenv("SAML_SP_METADATA_URL"),
			"SAML_REPLACE_DOMAIN_FROM": os.Getenv("SAML_REPLACE_DOMAIN_FROM"),
			"SAML_REPLACE_DOMAIN_TO":   os.Getenv("SAML_REPLACE_DOMAIN_TO"),
		},
	},
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
