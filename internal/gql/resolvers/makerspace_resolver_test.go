package resolvers

import (
	"context"
	"testing"
)

func TestMakerspaceStaff(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	wantedStaff := []int{data.Users[1].Id, data.Users[2].Id, data.Users[3].Id, data.Users[4].Id}
	gotStaff, err := resolver.Makerspace().Staff(context.TODO(), &data.Parlophone)
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
func TestMakerspaceManagers(t *testing.T) {
	_, data, resolver := helpersForResolverTest(t)
	wantedManagers := []int{data.Users[0].Id}
	gotManagers, err := resolver.Makerspace().Managers(context.TODO(), &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff of makerspace %v", err)
	}
	mIds := []int{}
	for _, user := range gotManagers {
		mIds = append(mIds, user.Id)
	}
	if !areGroupsEqual(wantedManagers, mIds) {
		t.Fatalf("manager group for makerspace not equal. Wanted %v but got %v", wantedManagers, mIds)
	}
}
