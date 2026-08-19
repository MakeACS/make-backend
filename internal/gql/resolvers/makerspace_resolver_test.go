package resolvers

import (
	"testing"
)

func TestMakerspaceStaffAuthed(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	ctx := ContextWithUser(t.Context(), data.Users[0].Id)
	wantedStaff := []int{data.Users[1].Id, data.Users[2].Id, data.Users[3].Id, data.Users[4].Id}
	gotStaff, err := resolver.Makerspace().Staff(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff of makerspace %v", err)
	}
	sIds := []int{}
	for _, user := range gotStaff {
		sIds = append(sIds, user.Id)
	}
	if !areGroupsEqual(wantedStaff, sIds) {
		t.Fatalf("staff group for makerspace not equal. Wanted %v but got %v", wantedStaff, sIds)
	}
}
func TestMakerspaceManagersAuthed(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	wantedManagers := []int{data.Users[0].Id}
	gotManagers, err := resolver.Makerspace().Managers(t.Context(), &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get managers of makerspace %v", err)
	}
	mIds := []int{}
	for _, user := range gotManagers {
		mIds = append(mIds, user.Id)
	}
	if !areGroupsEqual(wantedManagers, mIds) {
		t.Fatalf("manager group for makerspace not equal. Wanted %v but got %v", wantedManagers, mIds)
	}
}

func TestMakerspaceStaffSubgroups(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	ctx := ContextWithUser(t.Context(), data.Users[0].Id)

	wantedSG := []int{data.BeatlesMusicians.Id}
	gotSGs, err := resolver.Makerspace().StaffSubgroups(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff subgroups of makerspace %v", err)
	}
	SGIds := []int{}
	for _, user := range gotSGs {
		SGIds = append(SGIds, user.Id)
	}
	if !areGroupsEqual(wantedSG, SGIds) {
		t.Fatalf("staff subgroups for makerspace not equal. Wanted %v but got %v", wantedSG, SGIds)
	}
}

func TestMakerspaceManagerSubgroupsAuthed(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	ctx := ContextWithUser(t.Context(), data.Users[0].Id)

	wantedSG := []int{data.BeatlesManagers.Id}
	gotSGs, err := resolver.Makerspace().ManagerSubgroups(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get manager subgroups of makerspace: %v", err)
	}
	SGIds := []int{}
	for _, user := range gotSGs {
		SGIds = append(SGIds, user.Id)
	}
	if !areGroupsEqual(wantedSG, SGIds) {
		t.Fatalf("manager subgroups for makerspace not equal. Wanted %v but got %v", wantedSG, SGIds)
	}
}

func TestAddToStaffAgroupWithoutAuth(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	groupAlreadyThere := data.BeatlesMusicians.Id
	groupToAdd := data.ExMusicians.Id
	wantedSG := []int{groupAlreadyThere, groupToAdd}

	good, err := resolver.Mutation().SetAnonymousGroupSubgroups(t.Context(), data.Parlophone.StaffAgroupId, wantedSG)
	if good || err == nil {
		t.Fatalf("succeeded in adding subgroups to anonymous group when unauthenticated")
	}

}

func TestAddAndRemoveFromStaffAgroup(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	groupAlreadyThere := data.BeatlesMusicians.Id
	groupToAdd := data.ExMusicians.Id
	midWantedSG := []int{groupAlreadyThere, groupToAdd}
	endWantedSG := []int{groupAlreadyThere}

	ctx := ContextWithUser(t.Context(), data.Users[0].Id)

	good, err := resolver.Mutation().SetAnonymousGroupSubgroups(ctx, data.Parlophone.StaffAgroupId, midWantedSG)
	if !good || err != nil {
		t.Fatalf("failed to set agroup subgroups to %v: good: %v, err: %v", midWantedSG, good, err)
	}

	// add new group
	gotSGsMid, err := resolver.Makerspace().StaffSubgroups(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff subgroups of makerspace after addition: %v", err)
	}
	midSGIds := []int{}
	for _, user := range gotSGsMid {
		midSGIds = append(midSGIds, user.Id)
	}
	if !areGroupsEqual(midWantedSG, midSGIds) {
		t.Fatalf("failed to add subgroup to agroup. Wanted %v but got %v", midWantedSG, midSGIds)
	}

	// remove new group
	good, err = resolver.Mutation().SetAnonymousGroupSubgroups(ctx, data.Parlophone.StaffAgroupId, endWantedSG)
	if !good || err != nil {
		t.Fatalf("failed to set agroup subgroups to %v: good: %v, err: %v", endWantedSG, good, err)
	}

	gotSGsEnd, err := resolver.Makerspace().StaffSubgroups(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff subgroups of makerspace after removal: %v", err)
	}
	endSGIds := []int{}
	for _, user := range gotSGsEnd {
		endSGIds = append(endSGIds, user.Id)
	}
	if !areGroupsEqual(endWantedSG, endSGIds) {
		t.Fatalf("failed to remove subgroup from agroup. Wanted %v but got %v", endWantedSG, endSGIds)
	}
}
