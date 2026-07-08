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

func evalStaff(store *database.Store, ctx context.Context) (bool, error) {
	user_id, ok := ctx.Value("user_id").(int)

	if !ok {
		return false, fmt.Errorf("Could not extract user_id from context")
	}

	isStaff, err := store.Users.IsStaff(ctx, user_id)
	if err != nil {
		return false, err
	}

	return isStaff, nil
}

func isStaff(store *database.Store) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {

		staff, err := evalStaff(store, ctx)
		if err != nil {
			return nil, err
		}

		if staff {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
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

func evalSelf(ctx context.Context) (bool, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return false, fmt.Errorf("Failed to get field context")
	}

	target_id, ok := fc.Args["target_id"].(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from field context")
	}

	user_id := ctx.Value("user_id").(int)

	return target_id == user_id, nil
}

func isSelf() func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		self, err := evalSelf(ctx)
		if err != nil {
			return nil, err
		}

		if self {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}
}

func SetupDirectives(config *gql.Config, store *database.Store) {
	config.Directives.IsAdmin = isAdmin(store)
	config.Directives.IsManager = isManager(store)
	config.Directives.IsStaff = isStaff(store)
	config.Directives.IsTrainer = isTrainer(store)

	config.Directives.IsSelf = isSelf()
}
