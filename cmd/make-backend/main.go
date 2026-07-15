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
	"time"

	acsmqtt "make-backend/internal/api/acs/acs-mqtt"
	acsrest "make-backend/internal/api/rest"
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
	_ "github.com/lib/pq"
	"github.com/vektah/gqlparser/v2/ast"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	programLevel := &slog.LevelVar{} // Defaults to Info level
	opts := &slog.HandlerOptions{Level: programLevel}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))
	programLevel.Set(slog.LevelDebug)

}

const httpPort = 23003
const mqttPort = 23002

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

	// Database
	db, err := database.SetupDB()
	if err != nil {
		log.Fatalf("Failed to setup DB: %s", err)
	}
	defer db.Close()

	store := database.NewStore(db)
	logger := logging.NewLogger(store)

	httpServer := startHttp(db, store, logger, httpPort)
	mqttServer, _ := acsmqtt.StartMqtt(logger, store, mqttPort)
	reverseProxy := StartReverseProxy(port, httpPort, mqttPort)

	logger.AuditLog.CreateUnassociatedWithData("builtin.server.start.1", map[string]any{"time": time.Now()}, "Server started")

	<-done
	slog.Warn("caught signal, stopping...")
	reverseProxy.Close()
	mqttServer.Close()
	httpServer.Close()

	slog.Info("main.go finished")

}

func StartReverseProxy(port string, httpPort, mqttPort int) *http.Server {
	targetHttp, _ := url.Parse(fmt.Sprintf("http://localhost:%d", httpPort))
	targetMqtt, _ := url.Parse(fmt.Sprintf("http://localhost:%d", mqttPort))

	proxyHttp := httputil.NewSingleHostReverseProxy(targetHttp)
	proxyMqtt := httputil.NewSingleHostReverseProxy(targetMqtt)

	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/mqtt") {
			r.Host = targetMqtt.Host
			proxyMqtt.ServeHTTP(w, r)
		} else {
			r.Host = targetHttp.Host
			proxyHttp.ServeHTTP(w, r)
		}
	}
	server := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: http.HandlerFunc(handler)}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	return server

}

func startHttp(db *sql.DB, store *database.Store, logger *logging.Logger, port int) *http.Server {
	// Sessions
	sessionManager := auth.SetupSessions(db)
	// Auth
	samlMiddleware := auth.SetupSamlSP(store, sessionManager)

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

	mux.Handle("/saml/", samlMiddleware)

	protectedQueryHandler := sessionManager.LoadAndSave(auth.AuthContextMiddleware(srv, sessionManager))

	mux.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", protectedQueryHandler)

	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir("./client")))
	mux.Handle("/app/", fileHandler)
	mux.Handle("/", http.RedirectHandler("/app/", http.StatusFound))

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", port)

	acsrest.RegisterHandlers(mux)

	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", port),
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.Default().With("server", "http").Handler(), slog.LevelInfo)}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	return server
}
