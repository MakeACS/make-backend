package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"make-backend/internal/auth"
	"make-backend/internal/database"
	"make-backend/internal/gql"
	"make-backend/internal/logging"
	"make-backend/internal/plugins"
	auth_plugin "make-backend/internal/plugins/auth"
	"time"

	acsmqtt "make-backend/internal/api/acs/acs-mqtt"
	rest "make-backend/internal/api/rest"
	"make-backend/internal/gql/directives"
	"make-backend/internal/gql/resolvers"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexedwards/scs/v2"
	_ "github.com/lib/pq"
	"github.com/vektah/gqlparser/v2/ast"
)

//go:generate go run github.com/99designs/gqlgen generate

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	programLevel := &slog.LevelVar{} // Defaults to Info level
	opts := &slog.HandlerOptions{Level: programLevel}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))
	programLevel.Set(slog.LevelDebug)
}

const httpPort = 23003
const mqttPort = 23002

var glblPlugins = plugins.PluginStore{}

func GetAuthPlugin() auth_plugin.AuthProvider {
	for _, p := range glblPlugins.Auth {
		return p
	}
	return nil
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No PORT env found")
	}
	host := os.Getenv("HOST_URL")
	if port == "" {
		log.Fatal("No HOST_URL env found")
	}
	hostAndPort := fmt.Sprintf("%s:%s", host, port)

	// Database
	db, err := database.SetupDB()
	if err != nil {
		log.Fatalf("Failed to setup DB: %s", err)
	}
	defer db.Close()

	store := database.NewStore(db)
	logger := logging.NewLogger(store)
	// Sessions
	sessionManager := auth.SetupSessions(db)

	httpServer := startHttp(db, store, logger, httpPort, sessionManager)
	mqttServer, _ := acsmqtt.StartMqtt(logger, store, mqttPort)
	plugins, err := plugins.StartPlugins(hostAndPort, store, sessionManager)
	glblPlugins = plugins
	reverseProxy := StartReverseProxy(port, httpPort, mqttPort, plugins.HttpForwards)
	if err != nil {
		slog.Error("failed to start plugins", "err", err)
	}

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)

	logger.AuditLog.CreateUnassociatedWithData("builtin.server.start.1", map[string]any{"time": time.Now()}, "Server started")
	<-done
	slog.Warn("caught signal, stopping...")
	plugins.Shutdown()
	reverseProxy.Close()
	mqttServer.Close()
	httpServer.Close()

	slog.Info("main.go finished")

}

func StartReverseProxy(port string, httpPort, mqttPort int, pluginForwards []plugins.PluginHTTPForwarding) *http.Server {
	pluginRProxies := []*httputil.ReverseProxy{}
	for _, forward := range pluginForwards {
		target, _ := url.Parse(fmt.Sprintf("http://localhost:%d", forward.ToPort))

		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(target)

				inboundPath := pr.In.URL.Path
				if strings.HasPrefix(inboundPath, forward.Path) {
					pr.Out.URL.Path = strings.TrimPrefix(inboundPath, forward.Path)
				}
			},
		}
		pluginRProxies = append(pluginRProxies, proxy)
		slog.Info("Setting up plugin HTTP forwarding", "path", forward.Path, "to", target)
	}

	targetHttp, _ := url.Parse(fmt.Sprintf("http://localhost:%d", httpPort))
	targetMqtt, _ := url.Parse(fmt.Sprintf("http://localhost:%d", mqttPort))

	proxyHttp := httputil.NewSingleHostReverseProxy(targetHttp)
	proxyMqtt := httputil.NewSingleHostReverseProxy(targetMqtt)

	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/mqtt") {
			r.Host = targetMqtt.Host
			proxyMqtt.ServeHTTP(w, r)
			return
		}
		for i, path := range pluginForwards {
			if strings.HasPrefix(r.URL.Path, path.Path) {
				pluginRProxies[i].ServeHTTP(w, r)
				slog.Info("serving to plugin", "path", pluginForwards[i].Path, "to", pluginForwards[i].ToPort)
				return
			}
		}

		// default go to the http server
		r.Host = targetHttp.Host
		proxyHttp.ServeHTTP(w, r)

	}
	server := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: http.HandlerFunc(handler)}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	return server

}

func startHttp(db *sql.DB, store *database.Store, logger *logging.Logger, port int, sessionManager *scs.SessionManager) *http.Server {

	// GraphQL
	graphqlConfig := gql.Config{Resolvers: &resolvers.Resolver{Store: store}}
	directives.SetupDirectives(&graphqlConfig, store)
	srv := handler.New(gql.NewExecutableSchema(graphqlConfig))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	mux := http.NewServeMux()

	protectedQueryHandler := sessionManager.LoadAndSave(auth.RequiredAuthMiddleware(srv, sessionManager))

	mux.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", protectedQueryHandler)

	loginHandler := func(w http.ResponseWriter, r *http.Request) {
		a := GetAuthPlugin()

		req, err := a.GenerateLoginRequest(&auth_plugin.UserLoginStartRequest{
			OriginalURL: "http://localhost:8080",
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

	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir("./client")))
	mux.Handle("/app/", fileHandler)
	mux.HandleFunc("/login", loginHandler)

	mux.Handle("/", http.RedirectHandler("/app/", http.StatusFound))

	newAccountHandler := func(w http.ResponseWriter, r *http.Request) {
		userId := sessionManager.GetInt(r.Context(), "user_id")
		s := fmt.Sprintf("welcome. youre user id is %d. enter your name and stuff", userId)
		w.Write([]byte(s))
	}

	authedHandler := func(w http.ResponseWriter, r *http.Request) {
		userId := sessionManager.GetInt(r.Context(), "user_id")
		s := fmt.Sprintf("Hello user id %d", userId)
		w.Write([]byte(s))
	}
	mux.Handle("/app/newAccount", sessionManager.LoadAndSave(http.HandlerFunc(newAccountHandler)))
	mux.Handle("/app/authed", sessionManager.LoadAndSave(http.HandlerFunc(authedHandler)))

	rest.RegisterHandlers(mux)

	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", port),
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.Default().With("server", "http").Handler(), slog.LevelInfo)}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	return server
}
