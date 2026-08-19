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

type TestMockData struct {
	Users []struct {
		Id    int
		Email string
	}
	BeatlesMusicians models.Group
	OtherMusicians   models.Group
	ExMusicians      models.Group
	BeatlesManagers  models.Group
	Brits            models.Group

	Parlophone models.Makerspace
}

var localContext TestMockData = TestMockData{
	Users: []struct {
		Id    int
		Email string
	}{
		{Email: "brian@beatles.com"},
		{Email: "john@beatles.com"},
		{Email: "paul@beatles.com"},
		{Email: "george@beatles.com"},
		{Email: "ringo@beatles.com"},
		{Email: "pete@beatles.com"},
	},
}

func makeTestMakerspace(ctx context.Context, l *slog.Logger, store *Store) bool {
	id, err := store.Makerspaces.CreateMakerspace(ctx, "Parlophone", false)
	if err != nil {
		l.Error("Failed to create makerspace", "err", err)
		return false
	}
	m, err := store.Makerspaces.GetMakerspaceById(ctx, id)
	if err != nil {
		l.Error("Failed to get created makerspace", "err", err)
		return false
	}
	localContext.Parlophone = *m
	err = store.Groups.SetGroupsForAnonymousGroup(ctx, localContext.Parlophone.StaffAgroupId, []int{localContext.BeatlesMusicians.Id})
	if err != nil {
		l.Error("Failed to add beatles as staff of parlophone", "err", err)
		return false
	}
	err = store.Groups.SetGroupsForAnonymousGroup(ctx, localContext.Parlophone.ManagementAgroupId, []int{localContext.BeatlesManagers.Id})
	if err != nil {
		l.Error("Failed to add beatles managers as staff of parlophone", "err", err)
		return false
	}

	return true
}
func fillTestUsers(ctx context.Context, l *slog.Logger, store *Store) bool {
	for i, u := range localContext.Users {
		id, err := store.Users.CreateUser(ctx, u.Email)
		if err != nil {
			l.Error("failed to add test user", "err", err)
			return false
		}
		localContext.Users[i].Id = id
	}
	return true
}

func makeTestGroups(ctx context.Context, l *slog.Logger, store *Store) bool {
	managerGroupID, err := store.Groups.GetAdminGroupId(ctx)
	if err != nil {
		l.Error("Failed to find root group", "err", err)
		return false
	}
	localContext.BeatlesManagers, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   managerGroupID,
		Name:        "The Beatles Managers",
		Description: "",
	})
	if err != nil {
		l.Error("failed to create manager group", "err", err)
		return false
	}
	localContext.BeatlesMusicians, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   localContext.BeatlesManagers.Id,
		Name:        "The beatles musicians",
		Description: "the members of the beatles",
	})
	if err != nil {
		l.Error("failed to create musician group", "err", err)
		return false
	}
	localContext.OtherMusicians, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   localContext.BeatlesManagers.Id,
		Name:        "Other musicians",
		Description: "Other members the beatles don't know about",
	})
	if err != nil {
		l.Error("failed to create other musician group", "err", err)
		return false
	}

	localContext.ExMusicians, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   localContext.BeatlesManagers.Id,
		Name:        "The ex musicians for the beatles",
		Description: "ex members of the beatles",
	})
	if err != nil {
		l.Error("failed to create ex musician group", "err", err)
		return false
	}

	localContext.Brits, err = store.Groups.CreateGroup(ctx, models.Group{
		ManagerId:   managerGroupID,
		Name:        "British People",
		Description: "",
	})
	if err != nil {
		l.Error("failed to create british group", "err", err)
		return false
	}

	err = store.Groups.AddSubgroupToGroup(ctx, localContext.BeatlesMusicians.Id, localContext.Brits.Id, models.GroupViewPermission_SeeSelf)
	if err != nil {
		l.Error("Failed to add beatles subgroup to british super group")
		return false
	}

	err = store.Groups.AddSubgroupToGroup(ctx, localContext.BeatlesManagers.Id, localContext.Brits.Id, models.GroupViewPermission_SeeSelf)
	if err != nil {
		l.Error("Failed to add beatles managers subgroup to british super group")
		return false
	}

	return true
}

func addUsersToTestGroups(ctx context.Context, t *slog.Logger, store *Store) bool {
	for _, member := range []struct {
		email   string
		groupId int
		perm    models.GroupViewPermission
	}{
		{"brian@beatles.com", localContext.BeatlesManagers.Id, models.GroupViewPermission_SeeSelf},
		{"john@beatles.com", localContext.BeatlesMusicians.Id, models.GroupViewPermission_SeeAll},
		{"paul@beatles.com", localContext.BeatlesMusicians.Id, models.GroupViewPermission_SeeAll},
		{"george@beatles.com", localContext.BeatlesMusicians.Id, models.GroupViewPermission_SeeAll},
		{"ringo@beatles.com", localContext.BeatlesMusicians.Id, models.GroupViewPermission_SeeAll},
		{"pete@beatles.com", localContext.ExMusicians.Id, models.GroupViewPermission_SeeNone},
	} {
		user, err := store.Users.GetUserByEmail(ctx, member.email)
		if err != nil {
			t.Error("couldn't find user", "email", member.email, "err", err)
			return false
		}
		err = store.Groups.AddUserToGroup(ctx, user.Id, member.groupId, member.perm)
		if err != nil {
			t.Error("Failed to add ", "user", user, "group", member.groupId)
			return false
		}
	}

	return true

}

func fillTestData(ctx context.Context, l *slog.Logger, store *Store) {
	fillTestUsers(ctx, l, store)
	makeTestGroups(ctx, l, store)
	addUsersToTestGroups(ctx, l, store)
	makeTestMakerspace(ctx, l, store)

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

func VerifyTestDb(logger *slog.Logger) (*sql.DB, *Store, *TestMockData) {
	localLock.Lock()
	defer localLock.Unlock()
	if localDB != nil {
		return localDB, localStore, &localContext
	}
	var err error
	localDB, localCleanup, err = setupTestDB(logger)
	if err != nil {
		logger.Error("failed to setup test DB", "err", err)
	}
	localStore = NewStore(localDB)
	return localDB, localStore, &localContext
}
