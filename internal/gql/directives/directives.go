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

func isManager(store *database.Store) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		user_id, ok := ctx.Value("user_id").(int)

		if !ok {
			return nil, fmt.Errorf("Access denied")
		}

		isManager, err := store.Users.IsManager(ctx, user_id)
		if err != nil {
			return nil, err
		}

		return isManager, nil
	}
}

func isStaff(store *database.Store) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		user_id, ok := ctx.Value("user_id").(int)

		if !ok {
			return nil, fmt.Errorf("Access denied")
		}

		isStaff, err := store.Users.IsStaff(ctx, user_id)
		if err != nil {
			return nil, err
		}

		return isStaff, nil
	}
}

func isTrainer(store *database.Store) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		user_id, ok := ctx.Value("user_id").(int)

		if !ok {
			return nil, fmt.Errorf("Access denied")
		}

		isTrainer, err := store.Users.IsTrainer(ctx, user_id)
		if err != nil {
			return nil, err
		}

		return isTrainer, nil
	}
}

func SetupDirectives(config *gql.Config, store *database.Store) {
	config.Directives.IsAdmin = isAdmin(store)
	config.Directives.IsManager = isManager(store)
	config.Directives.IsStaff = isStaff(store)
	config.Directives.IsTrainer = isTrainer(store)
}
