package main

import (
	"fmt"
	"log"
	"log/slog"
	"make-backend/internal/auth"
	"make-backend/internal/database"
	"make-backend/internal/gql"
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
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/vektah/gqlparser/v2/ast"

	mqtt "github.com/mochi-mqtt/server/v2"
	mqttauth "github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	// Migrations
	log.Println("Running migrations...")
	goose.SetBaseFS(database.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %s", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}

	store := database.NewStore(db)

	httpServer := startHttp(store, httpPort)
	mqttServer := mqttSetup(mqttPort)
	reverseProxy := rproxySetup(port, httpPort, mqttPort)

	<-done
	slog.Warn("caught signal, stopping...")
	reverseProxy.Close()
	_ = mqttServer.Close()
	httpServer.Close()

	slog.Info("main.go finished")

}

func rproxySetup(port string, httpPort, mqttPort int) *http.Server {
	targetHttp, _ := url.Parse(fmt.Sprintf("http://localhost:%d", httpPort))
	targetMqtt, _ := url.Parse(fmt.Sprintf("http://localhost:%d", mqttPort))

	// Create proxy instances
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
func startHttp(store *database.Store, port int) *http.Server {
	// Auth
	samlMiddleware := auth.SetupSamlSP(store)

	// GraphQL
	srv := handler.New(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{Store: store}}))

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

	protectedQueryHandler := samlMiddleware.RequireAccount(
		auth.AuthContextMiddleware(srv),
	)
	mux.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", protectedQueryHandler)

	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir("./client")))
	mux.Handle("/app/", fileHandler)
	mux.Handle("/", http.RedirectHandler("/app/", http.StatusFound))

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", port)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	return server
}

func mqttSetup(port int) *mqtt.Server {
	wsCfg := listeners.Config{
		Type:    listeners.TypeWS,
		ID:      "std-listener",
		Address: fmt.Sprintf(":%v", port)}

	server := mqtt.New(nil)
	_ = server.AddHook(new(mqttauth.AllowHook), nil)

	wsListener := listeners.NewWebsocket(wsCfg)

	err := server.AddListener(wsListener)
	if err != nil {
		log.Fatalf("Failed to add listener: %v", err)
	}

	go server.Serve()
	return server
}
