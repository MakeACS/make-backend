package main

import (
	"log"
	"make-backend/internal/auth"
	"make-backend/internal/database"
	"make-backend/internal/gql"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
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

	http.Handle("/saml/", samlMiddleware)

	protectedQueryHandler := samlMiddleware.RequireAccount(
		auth.AuthContextMiddleware(srv),
	)

	http.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", protectedQueryHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
