package database

import (
	"log/slog"
	"testing"
)

func TestAddedUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db test in short mode.")
	}
	_, store := VerifyDb(slog.Default())
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
	_, store := VerifyDb(slog.Default())
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
		inGroup, _, err := store.Groups.IsUserInGroup(t.Context(), user.Id, localContext.beatlesMusicians.Id)
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
	_, store := VerifyDb(slog.Default())
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
		inGroup, _, err := store.Groups.IsUserInGroup(t.Context(), user.Id, localContext.brits.Id)
		if err != nil {
			t.Fatalf("couldn't check user in group for email '%s': %v", email, err)
		}
		if !inGroup {
			t.Fatalf("user with email '%s' should be in group '%s' via subgroup but wasn't", email, localContext.beatlesMusicians.Name)
		}
	}
}
