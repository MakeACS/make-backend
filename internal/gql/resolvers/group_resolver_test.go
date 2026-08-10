package resolvers

import (
	"context"
	"log/slog"
	"make-backend/internal/auth"
	"make-backend/internal/database"
	"testing"
)

func helpersForResolverTest(t *testing.T) (*database.Store, *database.TestMockData, *Resolver) {
	if testing.Short() {
		t.Skip("Skipping db/resolver test in short mode.")
	}
	_, store, mock := database.VerifyTestDb(slog.Default())
	return store, mock, &Resolver{Store: store}
}

func ContextWithUser(parent context.Context, userId int) context.Context {
	return context.WithValue(parent, auth.UserContextKey{}, userId)
}

// If you query a group for a member and you don't have visibility to that group, you should get a group not found error
func TestUserGroupVisibilityByMemberQueryGood(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)

	whenUserIsBrian := ContextWithUser(t.Context(), data.Users[0].Id)
	userId := 400

	_, brianErr := resolver.Query().IsUserInGroup(whenUserIsBrian, userId, data.OtherMusicians.Id)
	if brianErr != nil {
		t.Errorf("failed to check membership when brian should have visibility: %v", brianErr)
	}
}
func TestUserGroupVisibilityByMemberQueryBad(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)

	whenUserIsJohn := ContextWithUser(t.Context(), data.Users[1].Id)
	userId := 400

	_, JohnErr := resolver.Query().IsUserInGroup(whenUserIsJohn, userId, data.OtherMusicians.Id)
	if JohnErr == nil {
		t.Errorf("john checking membership should return a group not found error since he has no visibility to the other musicians but got null")
	}
}

func TestUserManagingGroup(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)

	trials := []struct {
		askerUser int
		user      int
		group     int

		shouldError  bool
		shouldManage bool
	}{
		{data.Users[0].Id, data.Users[0].Id, data.BeatlesMusicians.Id, false, true},
		{data.Users[1].Id, data.Users[1].Id, data.BeatlesMusicians.Id, false, false},
		{data.Users[1].Id, data.Users[1].Id, data.OtherMusicians.Id, true, false},
	}

	for _, trial := range trials {
		ctx := ContextWithUser(t.Context(), trial.askerUser)

		can, err := resolver.Query().CanUserManageGroup(ctx, trial.user, trial.group)
		if err != nil && !trial.shouldError {
			t.Fatalf("got error when %v checks if %v can manage group %v: %v", trial.askerUser, trial.user, trial.group, err)
		} else if err == nil && trial.shouldError {
			t.Fatalf("should get error when %v checks if %v can manage group %v we can't see", trial.askerUser, trial.user, trial.group)
			continue
		}
		if trial.shouldManage && can != trial.shouldManage {
			t.Fatalf("%v should be able to manage %v but wasn't able to for some reason", trial.user, trial.group)
		} else if !trial.shouldManage && can != trial.shouldManage {
			t.Fatalf("%v should NOT be able to manage %v but wasn't able to for some reason", trial.user, trial.group)
		}
	}
}

func TestGroupMembership(t *testing.T) {

	_, data, resolver := helpersForResolverTest(t)
	ctx := ContextWithUser(t.Context(), data.Users[0].Id)

	testCases := []struct {
		desc           string
		user           int
		group          int
		shouldBeMember bool
		shouldBeDirect bool
	}{
		{"brian as direct member of managers", data.Users[0].Id, data.BeatlesManagers.Id, true, true},
		{"brian as NOT member of beatles musicians", data.Users[0].Id, data.BeatlesMusicians.Id, false, false},
		{"john as direct member of beatles musicians", data.Users[1].Id, data.BeatlesMusicians.Id, true, true},
		{"john as indirect member of brits", data.Users[1].Id, data.Brits.Id, true, false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			inGroup, err1 := resolver.Query().IsUserInGroup(ctx, tC.user, tC.group)
			inGroupDirectly, err2 := resolver.Query().IsUserInGroup(ctx, tC.user, tC.group)
			if err1 != nil {
				t.Fatalf("failed to check if user %v in group %v: %v", tC.user, tC.group, err1)
			}
			if err2 != nil {
				t.Fatalf("failed to check if user %v directly in group %v: %v", tC.user, tC.group, err2)
			}
			if inGroup != tC.shouldBeMember {
				t.Fatalf("user %v should have membership:%v to group %v but was %v", tC.user, tC.shouldBeMember, tC.group, inGroup)
			}
			if inGroup != tC.shouldBeMember {
				t.Fatalf("user %v should have direct membership:%v to group %v but was %v", tC.user, tC.shouldBeDirect, tC.group, inGroupDirectly)
			}
		})
	}
}

func TestFindingGroupsById(t *testing.T) {

	_, data, resolver := helpersForResolverTest(t)
	testCases := []struct {
		desc        string
		askingUser  int
		group       int
		exists      bool
		shouldError bool
	}{
		{
			desc:        "brian getting beatles manager group",
			askingUser:  data.Users[0].Id,
			group:       data.BeatlesManagers.Id,
			exists:      true,
			shouldError: false,
		},
		{
			desc:        "brian getting beatles musician group",
			askingUser:  data.Users[0].Id,
			group:       data.BeatlesMusicians.Id,
			exists:      true,
			shouldError: false,
		},
		{
			desc:        "john getting beatles musician group",
			askingUser:  data.Users[1].Id,
			group:       data.BeatlesMusicians.Id,
			exists:      true,
			shouldError: false,
		},

		{
			desc:        "john getting beatles manager group",
			askingUser:  data.Users[1].Id,
			group:       data.BeatlesManagers.Id,
			exists:      true,
			shouldError: true,
		}, {
			desc:        "john getting nonexistent group",
			askingUser:  data.Users[1].Id,
			group:       -1,
			exists:      false,
			shouldError: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := ContextWithUser(t.Context(), tC.askingUser)
			g, err := resolver.Query().GroupByID(ctx, tC.group)
			if !tC.shouldError && err != nil {
				t.Fatalf("should not error when user %v gets group %v but got %v", tC.askingUser, tC.group, err)
				return
			}
			if !tC.exists && (err == nil || g != nil) {
				t.Fatalf("group %v didn't exist but was somehow successful in fetching. Got %+v", tC.group, g)
				return
			}

			if tC.shouldError && err == nil {
				t.Fatalf("should error when user %v gets group %v but didn't", tC.askingUser, tC.group)
				return
			}

			if !tC.shouldError {
				if tC.exists && g == nil {
					t.Fatalf("group %v should exist but could not find it: err = %v", tC.group, err)
					return
				}
				if tC.group != g.Id {
					t.Fatalf("asked for group %v but got non-matching %+v", tC.group, g)
					return
				}
			}
		})
	}
}
