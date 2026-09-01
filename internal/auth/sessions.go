package auth

import (
	"database/sql"
	"log/slog"
	"make-backend/internal/plugins"
	auth_plugin "make-backend/internal/plugins/auth"

	"net/http"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

func SetupSessions(db *sql.DB) *scs.SessionManager {

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	return sessionManager
}

func LoginHandler(plugins *plugins.PluginStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vs := r.URL.Query()
		pluginId := vs.Get("plugin")
		if pluginId == "" {
			// no plugin specified. should redirect to login chooser page
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fromUrl := vs.Get("from")
		if fromUrl == "" {
			fromUrl = "http://localhost:8080"
		}

		a, ok := plugins.Auth[pluginId]
		if !ok {
			// no plugin found. should redirect to login chooser page
			w.WriteHeader(http.StatusNotFound)
			return
		}

		req, err := a.Provider.GenerateLoginRequest(&auth_plugin.UserLoginStartRequest{
			OriginalURL: fromUrl,
		})
		if err != nil {
			slog.Error("failed to get login url", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h := w.Header()
		for _, header := range req.SetHeaders {
			h.Add(header.Key, header.Value)
		}
		w.WriteHeader(int(req.Code))
		_, err = w.Write(req.Body)
		if err != nil {
			slog.Error("failed to write login redirect", "err", err)
		}
	}
}
