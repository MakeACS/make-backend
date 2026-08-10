package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func SetupDB() (*sql.DB, error) {
	db, err := createDB()
	if err != nil {
		return db, err
	}
	err = migrateDB(db)
	if err != nil {
		return db, err
	}
	return db, nil
}

func createDB() (*sql.DB, error) {
	db_user, db_pass, db_host, db_port, db_name := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")

	db_url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", db_user, db_pass, db_host, db_port, db_name)

	// setup postgresql connection pool
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(db *sql.DB) error {
	// Migrations
	log.Println("Running migrations...")
	goose.SetBaseFS(EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("Failed to run migrations: %w", err)
	}
	return nil
}
