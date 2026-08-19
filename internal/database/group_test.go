package database

import (
	"log/slog"
	"testing"
)

func TestUserGroupVisibilityGood(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store, data := VerifyTestDb(slog.Default())
	user := data.Users[0]
	group := data.BeatlesMusicians

	// brian can see musicians
	visible, err := store.Groups.IsGroupVisibleToUser(t.Context(), user.Id, group.Id)
	if err != nil {
		t.Fatalf("Failed to check group visibility of %+v to see %+v: %v", user, group, err)
	}
	if visible != true {
		t.Fatalf("%v should be able to see %v but couldn't", user, group)
	}
}
func TestUserGroupVisibilityBad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store, data := VerifyTestDb(slog.Default())
	user := data.Users[1]
	group := data.BeatlesManagers

	// musicians can't see up (unless explicitly allowed to)
	visible, err := store.Groups.IsGroupVisibleToUser(t.Context(), user.Id, group.Id)
	if err != nil {
		t.Fatalf("Failed to check group visibility of %+v to see %+v: %v", user, group, err)
	}
	if visible != false {
		t.Fatalf("%v should NOT be able to see %v but could", user, group)
	}
}

func TestAddedUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store, _ := VerifyTestDb(slog.Default())
	for _, user := range localContext.Users {
		user, err := store.Users.GetUserByEmail(t.Context(), user.Email)
		if err != nil {
			t.Fatalf("Failed to find user with email '%s': %v", user.Email, err)
		}
	}
}

func TestGroupMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store, _ := VerifyTestDb(slog.Default())
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
		inGroup, _, err := store.Groups.IsUserInGroup(t.Context(), user.Id, localContext.BeatlesMusicians.Id)
		if err != nil {
			t.Fatalf("couldn't check user in group for email '%s': %v", email, err)
		}
		if !inGroup {
			t.Fatalf("user with email '%s' should be in group '%s' but wasn't", email, localContext.BeatlesMusicians.Name)
		}
	}
}

func TestSubgroupMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store, _ := VerifyTestDb(slog.Default())
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
		inGroup, _, err := store.Groups.IsUserInGroup(t.Context(), user.Id, localContext.Brits.Id)
		if err != nil {
			t.Fatalf("couldn't check user in group for email '%s': %v", email, err)
		}
		if !inGroup {
			t.Fatalf("user with email '%s' should be in group '%s' via subgroup but wasn't", email, localContext.BeatlesMusicians.Name)
		}
	}
}

func TestCanGroupManageAnonymousGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store, data := VerifyTestDb(slog.Default())
	rootGroupId, err := store.Groups.GetAdminGroupId(t.Context())
	if err != nil {
		t.Fatalf("couldn't get root group ID to continue test")
	}
	tests := []struct {
		testName string
		// params
		managerGroupId   int
		anonymousGroupId int
		want             bool
		wantErr          bool
	}{
		{
			"beatles managers manage parlophone so can manage staff of parlophone",
			data.BeatlesManagers.Id,
			data.Parlophone.StaffAgroupId,
			true,
			false,
		},
		{
			"beatles musicians DONT manage parlophone so CANT manage staff of parlophone",
			data.BeatlesMusicians.Id,
			data.Parlophone.StaffAgroupId,
			false,
			false,
		},
		{
			"root group can manage any agroup so can manage staff of parlophone",
			rootGroupId,
			data.Parlophone.StaffAgroupId,
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, gotErr := store.Groups.CanGroupManageAnonymousGroup(t.Context(), tt.managerGroupId, tt.anonymousGroupId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CanGroupManageAnonymousGroup() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CanGroupManageAnonymousGroup() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("CanGroupManageAnonymousGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
