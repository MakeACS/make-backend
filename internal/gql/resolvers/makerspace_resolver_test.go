package resolvers

import (
	"context"
	"make-backend/internal/gql"
	"make-backend/internal/gql/directives"
	"net/http"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"
)

func gqlHandler(t *testing.T) http.Handler {
	store, _, _ := helpersForResolverTest(t)

	graphqlConfig := gql.Config{Resolvers: &Resolver{Store: store}}
	directives.SetupDirectives(&graphqlConfig, store)
	srv := handler.New(gql.NewExecutableSchema(graphqlConfig))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	return srv
}

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

func addUserIDContext(userId int) client.Option {
	return func(req *client.Request) {
		ctx := ContextWithUser(context.Background(), userId)
		req.HTTP = req.HTTP.WithContext(ctx)
	}
}

func TestMakerspaceStaffSubgroups(t *testing.T) {

	// _, data, _ := helpersForResolverTest(t)
	// srv := gqlHandler(t)
	// c := client.New(srv, addUserIDContext(data.Users[0].Id))
	// e := c.Post("")
	// wantedSG := []int{data.BeatlesMusicians.Id}
	// ag, err := resolver.Makerspace().StaffAnonymousGroup(ctx, &data.Parlophone)
	// //
	// if err != nil {
	// 	t.Fatalf("failed to get staff subgroups of makerspace %v", err)
	// }
	// SGIds := []int{}
	// for _, sg := range gotSGs {
	// 	SGIds = append(SGIds, sg.Id)
	// }
	// if !areGroupsEqual(wantedSG, SGIds) {
	// 	t.Fatalf("staff subgroups for makerspace not equal. Wanted %v but got %v", wantedSG, SGIds)
	// }
}

func TestMakerspaceManagerSubgroupsAuthed(t *testing.T) {
	// TODO: rewrite with gql client: calling resolvers bypasses directives which provide our auth
	// _, data, resolver := helpersForResolverTest(t)
	// ctx := ContextWithUser(t.Context(), data.Users[0].Id)
	//
	// wantedSG := []int{data.BeatlesManagers.Id}
	// gotSGs, err := resolver.Makerspace().ManagerSubgroups(ctx, &data.Parlophone)
	// if err != nil {
	// t.Fatalf("failed to get manager subgroups of makerspace: %v", err)
	// }
	// SGIds := []int{}
	// for _, user := range gotSGs {
	// SGIds = append(SGIds, user.Id)
	// }
	// if !areGroupsEqual(wantedSG, SGIds) {
	// t.Fatalf("manager subgroups for makerspace not equal. Wanted %v but got %v", wantedSG, SGIds)
	// }
}

func TestAddToStaffAgroupWithoutAuth(t *testing.T) {
	// TODO: rewrite with gql client: calling resolvers bypasses directives which provide our auth
	// _, data, resolver := helpersForResolverTest(t)
	// groupAlreadyThere := data.BeatlesMusicians.Id
	// groupToAdd := data.ExMusicians.Id
	// wantedSG := []int{groupAlreadyThere, groupToAdd}
	//
	// good, err := resolver.Mutation().SetStaffSubgroups(t.Context(), data.Parlophone.Id, wantedSG)
	// if good || err == nil {
	// t.Fatalf("succeeded in adding subgroups to anonymous group without user signed in. SHOULD DENY")
	// }
}

func TestAddToStaffAgroupWithWrongAuth(t *testing.T) {
	// TODO: rewrite with gql client: calling resolvers bypasses directives which provide our auth
	// _, data, resolver := helpersForResolverTest(t)
	// ctx := ContextWithUser(t.Context(), data.Users[4].Id)
	//
	// groupAlreadyThere := data.BeatlesMusicians.Id
	// groupToAdd := data.ExMusicians.Id
	// wantedSG := []int{groupAlreadyThere, groupToAdd}
	//
	// good, err := resolver.Mutation().SetStaffSubgroups(ctx, data.Parlophone.Id, wantedSG)
	// if good || err == nil {
	// t.Fatalf("succeeded in adding subgroups to anonymous group when user does not have that permission. SHOULD DENY")
	// }
}

func TestAddAndRemoveFromStaffAgroup(t *testing.T) {
	// TODO: rewrite with gql client: calling resolvers bypasses directives which provide our auth
	_, data, resolver := helpersForResolverTest(t)
	groupAlreadyThere := data.BeatlesMusicians.Id
	groupToAdd := data.ExMusicians.Id
	midWantedSG := []int{groupAlreadyThere, groupToAdd}
	endWantedSG := []int{groupAlreadyThere}

	ctx := ContextWithUser(t.Context(), data.Users[0].Id)

	good, err := resolver.Mutation().SetStaffSubgroups(ctx, data.Parlophone.Id, midWantedSG)
	if !good || err != nil {
		t.Fatalf("failed to set agroup subgroups to %v: good: %v, err: %v", midWantedSG, good, err)
	}

	// add new group
	ag, err := resolver.Makerspace().StaffAnonymousGroup(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff anonymous group after addition: %v", err)
	}
	gotSGsMid, err := resolver.AnonymousGroup().Subgroups(ctx, ag)
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
	good, err = resolver.Mutation().SetStaffSubgroups(ctx, data.Parlophone.Id, endWantedSG)
	if !good || err != nil {
		t.Fatalf("failed to set agroup subgroups to %v: good: %v, err: %v", endWantedSG, good, err)
	}

	ag, err = resolver.Makerspace().StaffAnonymousGroup(ctx, &data.Parlophone)
	if err != nil {
		t.Fatalf("failed to get staff anonymous group after removal: %v", err)
	}
	gotSGsEnd, err := resolver.AnonymousGroup().Subgroups(ctx, ag)
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
