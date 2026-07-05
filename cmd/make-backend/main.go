package main

import (
	"database/sql"
	"fmt"
	"log"
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
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env: %s", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No PORT env found")
	}

	db_user, db_pass, db_host, db_port, db_name := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")

	db_url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", db_user, db_pass, db_host, db_port, db_name)

	// setup postgresql connection pool
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := database.NewStore(db)

	srv := handler.New(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{Store: store}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
