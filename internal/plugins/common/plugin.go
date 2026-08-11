package common

type PluginType string

var (
	PluginType_Auth             PluginType = "auth"
	PluginType_CurrencyProvider PluginType = "currency_provider"
	PluginType_CurrencyConsumer PluginType = "currency_consumer"
	PluginType_Notification     PluginType = "notification"
)

type PluginInfo struct {
	Id    string
	About string
	Port  int16 // 0 if the plugin does not require forwarding http requests

}

var MagicKey string = "ACS_PLUGIN_ID"

type PluginDescription struct {
	Name        string
	PluginType  PluginType
	Version     uint
	MagicValue  string
	DownloadUrl string /// empty url implies that the plugin will just be there and theres no need to download
}
