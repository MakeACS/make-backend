package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupTestDB() (*sql.DB, func(), error) {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "test_user"
	dbPassword := "test_password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		// postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		// postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	deferer := func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return nil, func() {}, err
	}

	connStr := postgresContainer.MustConnectionString(ctx, "sslmode=disable")
	slog.Info("connecting to test postgres", "conn", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		return nil, deferer, err
	}

	slog.Info("Pinged DB")
	err = migrateDB(db)
	if err != nil {
		return db, deferer, err
	}

	return db, deferer, nil

}

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
