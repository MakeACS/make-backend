package main

import (
	"database/sql"
	"fmt"
	"log"
	"make-backend/internal/database"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	godotenv.Load()

	db_user, db_pass, db_host, db_port, db_name := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")

	db_url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", db_user, db_pass, db_host, db_port, db_name)

	// setup postgresql connection pool
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	goose.SetBaseFS(database.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %s", err)
	}

	if err := goose.Down(db, "migrations"); err != nil {
		log.Fatalf("Failed to migrate down: %s", err)
	}
}
