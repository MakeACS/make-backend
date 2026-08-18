package plugins

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"make-backend/internal/database"
	"make-backend/internal/plugins/auth"
	"make-backend/internal/plugins/common"
	"make-backend/internal/plugins/notify"
	"net/http"
	"os/exec"
	"path"

	"github.com/alexedwards/scs/v2"
	"github.com/hashicorp/go-plugin"
)

type TestAuthCBProvider struct {
	sessionManager *scs.SessionManager
	store          *database.Store
}

// UserLoggedIn implements [auth.AuthCallbackProvider].
func (t *TestAuthCBProvider) UserLoggedIn(data *auth.UserLoginCallback) (*auth.RedirectURL, error) {
	// create a new session and inform plugin how to inform client what to do next
	// {set headers, redirect}
	ret := auth.RedirectURL{}
	slog.Info("User logged in", "data", data)

	var newAccount bool = false
	u, err := t.store.Users.GetUserByEmail(context.TODO(), data.Email)
	if errors.Is(err, sql.ErrNoRows) {
		var id int
		id, err = t.store.Users.CreateUser(context.TODO(), data.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to create new user on login {should make new user}: %w", err)
		} else {
			slog.Info("new user joined", "email", data.Email)
			newAccount = true
		}
		u, err = t.store.Users.GetUserById(context.TODO(), id)
	}
	if err != nil || u == nil {
		return nil, fmt.Errorf("failed to find user on login {should make new user}: %w", err)
	}

	ctx1 := context.Background()
	ctx2, err := t.sessionManager.Load(ctx1, "")

	t.sessionManager.Put(ctx2, "user_id", u.Id)

	if err != nil {
		slog.Warn("failed to load token", "err", err)
		return nil, err
	}
	token, expiry, err := t.sessionManager.Commit(ctx2)
	if err != nil {
		slog.Warn("failed to commit token", "err", err)
		return nil, err
	}
	w := common.WriteAdapter{
		Headers: http.Header{},
		Body:    bytes.Buffer{},
		Code:    0,
	}

	t.sessionManager.WriteSessionCookie(ctx2, &w, token, expiry)

	_, headers, _ := w.IntoParts()
	if newAccount {
		ret.GotoUrl = "http://localhost:8080/app/newAccount"
	} else {
		ret.GotoUrl = "http://localhost:8080/app/authed"
	}
	for k, vs := range headers {
		for _, v := range vs {
			ret.SetHeaders = append(ret.SetHeaders, &auth.SetKV{
				Key:   k,
				Value: v,
			})
		}
	}

	return &ret, nil
}

// UserLoggedOut implements [auth.AuthCallbackProvider].
func (t *TestAuthCBProvider) UserLoggedOut(*auth.UserLogOffRequest) error {
	// end a session and inform plugin how to inform client what to do next
	// {remove headers, redirect}
	// maybe not needed since we probably don't support IDP side logout
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

type PluginStore struct {
	Notification map[string]notify.NotificationProvider
	Auth         map[string]auth.AuthProvider

	HttpForwards []PluginHTTPForwarding
}

func (PluginStore) Shutdown() {
	plugin.CleanupClients()
}

func StartPlugins(host string, store *database.Store, sessionManager *scs.SessionManager) (PluginStore, error) {
	plugin_dir := path.Join("./plugins", "bin")

	forwards := []PluginHTTPForwarding{}
	var plugins = PluginStore{
		Notification: map[string]notify.NotificationProvider{},
		Auth:         map[string]auth.AuthProvider{},
	}

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

		switch plugin_desc.PluginType {
		case common.PluginType_Auth:
			authPlugin, ok := raw.(auth.AuthProvider)
			if !ok {
				slog.Warn("plugin lied about type", "wanted", plugin_desc.PluginType, "plugin", plugin_desc.Name)
			}
			authPlugin.RegisterCallbackProvider(&TestAuthCBProvider{sessionManager, store})
			plugins.Auth[plugin_desc.Name] = authPlugin

		case common.PluginType_Notification:
			notifyPlugin, ok := raw.(notify.NotificationProvider)
			if !ok {
				slog.Warn("plugin lied about type", "wanted", plugin_desc.PluginType, "plugin", plugin_desc.Name)
			}
			plugins.Notification[plugin_desc.Name] = notifyPlugin

		}

	}
	plugins.HttpForwards = forwards

	return plugins, nil

}
