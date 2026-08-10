package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"make-backend/internal/database/models"
	"sync"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type testContext struct {
	users []struct {
		id    int
		email string
	}
	beatlesMusicians models.Group
	beatlesManagers  models.Group
	brits            models.Group
}

var localContext testContext = testContext{
	users: []struct {
		id    int
		email string
	}{
		{email: "brian@beatles.com"},
		{email: "john@beatles.com"},
		{email: "paul@beatles.com"},
		{email: "george@beatles.com"},
		{email: "ringo@beatles.com"},
	},
}

func fillTestUsers(ctx context.Context, l *slog.Logger, store *Store) bool {
	for i, u := range localContext.users {
		id, err := store.Users.CreateUser(ctx, u.email)
		if err != nil {
			l.Error("failed to add test user", "err", err)
			return false
		}
		localContext.users[i].id = id
	}
	return true
}

func makeTestGroups(ctx context.Context, l *slog.Logger, store *Store) bool {
	managerGroupID, err := store.Groups.GetAdminGroupId(ctx)
	if err != nil {
		l.Error("Failed to find root group", "err", err)
		return false
	}
	localContext.beatlesManagers, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   managerGroupID,
		Name:        "The Beatles Managers",
		Description: "",
	})
	if err != nil {
		l.Error("failed to create manager group", "err", err)
		return false
	}
	localContext.beatlesMusicians, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   localContext.beatlesManagers.Id,
		Name:        "The beatles musicians",
		Description: "the members of the beatles",
	})
	if err != nil {
		l.Error("failed to create musician group", "err", err)
		return false
	}

	localContext.brits, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   managerGroupID,
		Name:        "British People",
		Description: "",
	})
	if err != nil {
		l.Error("failed to create british group", "err", err)
		return false
	}

	err = store.Groups.AddSubgroupToGroup(ctx, localContext.beatlesMusicians.Id, localContext.brits.Id, models.GroupViewPermission_SeeSelf)
	if err != nil {
		l.Error("Failed to add beatles subgroup to british super group")
		return false
	}

	err = store.Groups.AddSubgroupToGroup(ctx, localContext.beatlesManagers.Id, localContext.brits.Id, models.GroupViewPermission_SeeSelf)
	if err != nil {
		l.Error("Failed to add beatles managers subgroup to british super group")
		return false
	}
	return true
}

func addUsersToTestGroups(ctx context.Context, t *slog.Logger, store *Store) bool {
	for _, email := range []string{
		"john@beatles.com",
		"paul@beatles.com",
		"george@beatles.com",
		"ringo@beatles.com",
	} {
		user, err := store.Users.GetUserByEmail(ctx, email)
		if err != nil {
			t.Error("couldn't find user", "email", email, "err", err)
			return false
		}
		err = store.Groups.AddUserToGroup(ctx, user.Id, localContext.beatlesMusicians.Id, models.GroupViewPermission_SeeAll)
		if err != nil {
			t.Error("Failed to add ", "user", user, "group", localContext.beatlesMusicians)
			return false
		}
	}

	user, err := store.Users.GetUserByEmail(ctx, "brian@beatles.com")
	if err != nil {
		t.Error("couldn't find user", "email", "brian@beatles.com", "err", err)
		return false
	}
	err = store.Groups.AddUserToGroup(ctx, user.Id, localContext.beatlesManagers.Id, models.GroupViewPermission_SeeAll)
	if err != nil {
		t.Error("Failed to add to group", "user", user, "group", localContext.beatlesManagers, "err", err)
		return false
	}
	return true

}

func fillTestData(ctx context.Context, l *slog.Logger, store *Store) {
	fillTestUsers(ctx, l, store)
	makeTestGroups(ctx, l, store)
	addUsersToTestGroups(ctx, l, store)

}

func setupTestDB(logger *slog.Logger) (*sql.DB, func(), error) {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "test_user"
	dbPassword := "test_password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	deferer := func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			logger.Error("failed to terminate container", "err", err)
		}
	}
	if err != nil {
		logger.Error("failed to start container", "err", err)
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

	err = migrateTestDB(logger, db)
	if err != nil {
		return db, deferer, err
	}

	store := NewStore(db)
	fillTestData(context.TODO(), logger, store)

	return db, deferer, nil

}

//go:embed migrations/*.sql
var testEmbedMigrations embed.FS

func migrateTestDB(logger *slog.Logger, db *sql.DB) error {
	// Migrations
	logger.Info("Running migrations...")
	goose.SetBaseFS(testEmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

var localLock sync.Mutex
var localDB *sql.DB = nil
var localStore *Store = nil
var localCleanup func() = func() {}

func VerifyDb(logger *slog.Logger) (*sql.DB, *Store) {
	localLock.Lock()
	defer localLock.Unlock()
	if localDB != nil {
		return localDB, localStore
	}
	var err error
	localDB, localCleanup, err = setupTestDB(logger)
	if err != nil {
		logger.Error("failed to setup test DB", "err", err)
	}
	localStore = NewStore(localDB)
	return localDB, localStore
}
