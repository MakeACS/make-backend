package directives

import (
	"context"
	"fmt"
	"make-backend/internal/database"
	"make-backend/internal/gql"

	"github.com/99designs/gqlgen/graphql"
)

func isAdmin(store *database.Store) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		user_id, ok := ctx.Value("user_id").(int)

		if !ok {
			return nil, fmt.Errorf("Access denied")
		}

		user, err := store.Users.GetUserById(ctx, user_id)
		if err != nil {
			return nil, err
		}

		return user.Admin, nil
	}
}

func SetupDirectives(config *gql.Config, store *database.Store) {
	config.Directives.IsAdmin = isAdmin(store)

}
