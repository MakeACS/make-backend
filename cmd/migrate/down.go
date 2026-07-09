package main

import (
	"log"
	"make-backend/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	godotenv.Load()

	db, err := database.SetupDB()
	if err != nil {
		log.Fatalf("Failed to setup DB: %s", err)
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
