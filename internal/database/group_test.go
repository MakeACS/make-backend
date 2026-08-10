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
	"testing"

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

func setupTestDB(t *testing.T) (*sql.DB, func(), error) {
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
			t.Logf("failed to terminate container: %s", err)
		}
	}
	if err != nil {
		t.Logf("failed to start container: %s", err)
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

	err = migrateTestDB(t, db)
	if err != nil {
		return db, deferer, err
	}

	store := NewStore(db)
	fillTestData(t, store)

	return db, deferer, nil

}

//go:embed migrations/*.sql
var testEmbedMigrations embed.FS

func migrateTestDB(t *testing.T, db *sql.DB) error {
	// Migrations
	t.Log("Running migrations...")
	goose.SetBaseFS(testEmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("Failed to run migrations: %w", err)
	}
	return nil
}

func fillTestUsers(t *testing.T, store *Store) {
	for i, u := range localContext.users {
		id, err := store.Users.CreateUser(t.Context(), u.email)
		if err != nil {
			t.Fatal("failed to add test user", err)
		}
		localContext.users[i].id = id
	}
}

func makeTestGroups(t *testing.T, store *Store) {
	managerGroupID, err := store.Groups.GetAdminGroupId(t.Context())
	if err != nil {
		t.Fatal("Failed to find root group", err)
	}
	localContext.beatlesManagers, err = store.Groups.CreateGroup(t.Context(), models.Group{
		ManagerId:   managerGroupID,
		Name:        "The Beatles Managers",
		Description: "",
	})
	if err != nil {
		t.Fatal("failed to create manager group", err)
	}
	localContext.beatlesMusicians, err = store.Groups.CreateGroup(t.Context(), models.Group{
		ManagerId:   localContext.beatlesManagers.Id,
		Name:        "The beatles musicians",
		Description: "the members of the beatles",
	})
	if err != nil {
		t.Fatal("failed to create musician group", err)
	}

	localContext.brits, err = store.Groups.CreateGroup(t.Context(), models.Group{
		ManagerId:   managerGroupID,
		Name:        "British People",
		Description: "",
	})
	if err != nil {
		t.Fatal("failed to create british group", err)
	}

	err = store.Groups.AddSubgroupToGroup(t.Context(), localContext.beatlesMusicians.Id, localContext.brits.Id, models.GroupViewPermission_SeeSelf)
	if err != nil {
		t.Fatal("Failed to add beatles subgroup to british super group")
	}

	err = store.Groups.AddSubgroupToGroup(t.Context(), localContext.beatlesManagers.Id, localContext.brits.Id, models.GroupViewPermission_SeeSelf)
	if err != nil {
		t.Fatal("Failed to add beatles managers subgroup to british super group")
	}
}

func addUsersToTestGroups(t *testing.T, store *Store) {
	for _, email := range []string{
		"john@beatles.com",
		"paul@beatles.com",
		"george@beatles.com",
		"ringo@beatles.com",
	} {
		user, err := store.Users.GetUserByEmail(t.Context(), email)
		if err != nil {
			t.Fatalf("couldn't find user with email '%s': %v", email, err)
		}
		err = store.Groups.AddUserToGroup(t.Context(), user.Id, localContext.beatlesMusicians.Id, models.GroupViewPermission_SeeAll)
		if err != nil {
			t.Fatalf("Failed to add %+v to group %+v", user, localContext.beatlesMusicians)
		}
	}

	user, err := store.Users.GetUserByEmail(t.Context(), "brian@beatles.com")
	if err != nil {
		t.Fatalf("couldn't find user with email '%s': %v", "brian@beatles.com", err)
	}
	err = store.Groups.AddUserToGroup(t.Context(), user.Id, localContext.beatlesManagers.Id, models.GroupViewPermission_SeeAll)
	if err != nil {
		t.Fatalf("Failed to add %+v to group %+v", user, localContext.beatlesManagers)
	}

}

func fillTestData(t *testing.T, store *Store) {
	fillTestUsers(t, store)
	makeTestGroups(t, store)
	addUsersToTestGroups(t, store)

}

var localLock sync.Mutex
var localDB *sql.DB = nil
var localStore *Store = nil
var localCleanup func() = func() {}

func verifyDb(t *testing.T) (*sql.DB, *Store) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
		return nil, nil
	}
	localLock.Lock()
	defer localLock.Unlock()
	if localDB != nil {
		return localDB, localStore
	}
	var err error
	localDB, localCleanup, err = setupTestDB(t)
	if err != nil {
		t.Fatal("Failed to setup test DB", err)
	}
	localStore = NewStore(localDB)
	return localDB, localStore
}

func TestAddedUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store := verifyDb(t)
	for _, user := range localContext.users {
		user, err := store.Users.GetUserByEmail(t.Context(), user.email)
		if err != nil {
			t.Fatalf("Failed to find user with email '%s': %v", user.Email, err)
		}
	}
}

func TestGroupMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store := verifyDb(t)
	for _, email := range []string{
		"john@beatles.com",
		"paul@beatles.com",
		"george@beatles.com",
		"ringo@beatles.com",
	} {
		user, err := store.Users.GetUserByEmail(t.Context(), email)
		if err != nil {
			t.Fatalf("couldn't find user with email '%s': %v", email, err)
		}
		inGroup, err := store.Groups.IsUserInGroup(t.Context(), user.Id, localContext.beatlesMusicians.Id)
		if err != nil {
			t.Fatalf("couldn't check user in group for email '%s': %v", email, err)
		}
		if !inGroup {
			t.Fatalf("user with email '%s' should be in group '%s' but wasn't", email, localContext.beatlesMusicians.Name)
		}
	}
}

func TestSubgroupMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store := verifyDb(t)
	for _, email := range []string{
		"john@beatles.com",
		"paul@beatles.com",
		"george@beatles.com",
		"ringo@beatles.com",
		"brian@beatles.com",
	} {
		user, err := store.Users.GetUserByEmail(t.Context(), email)
		if err != nil {
			t.Fatalf("couldn't find user with email '%s': %v", email, err)
		}
		inGroup, err := store.Groups.IsUserInGroup(t.Context(), user.Id, localContext.brits.Id)
		if err != nil {
			t.Fatalf("couldn't check user in group for email '%s': %v", email, err)
		}
		if !inGroup {
			t.Fatalf("user with email '%s' should be in group '%s' via subgroup but wasn't", email, localContext.beatlesMusicians.Name)
		}
	}
}
