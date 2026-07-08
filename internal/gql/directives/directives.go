package directives

import (
	"context"
	"fmt"
	"make-backend/internal/database"
	"make-backend/internal/gql"

	"github.com/99designs/gqlgen/graphql"
)

func isAdmin(store *database.Store, ctx context.Context) (bool, error) {
	user_id, ok := ctx.Value("user_id").(int)

	if !ok {
		return false, fmt.Errorf("Access denied")
	}

	user, err := store.Users.GetUserById(ctx, user_id)
	if err != nil {
		return false, err
	}

	return user.Admin, nil
}

func isManager(store *database.Store, ctx context.Context) (bool, error) {
	user_id, ok := ctx.Value("user_id").(int)

	if !ok {
		return false, fmt.Errorf("Failed to get user_id from context")
	}

	manager, err := store.Users.IsManager(ctx, user_id)
	if err != nil {
		return false, err
	}

	return manager, nil
}

func isStaff(store *database.Store, ctx context.Context) (bool, error) {
	user_id, ok := ctx.Value("user_id").(int)

	if !ok {
		return false, fmt.Errorf("Could not extract user_id from context")
	}

	staff, err := store.Users.IsStaff(ctx, user_id)
	if err != nil {
		return false, err
	}

	return staff, nil
}

func isTrainer(store *database.Store, ctx context.Context) (bool, error) {
	user_id, ok := ctx.Value("user_id").(int)

	if !ok {
		return false, fmt.Errorf("Failed to get user_id from context")
	}

	trainer, err := store.Users.IsTrainer(ctx, user_id)
	if err != nil {
		return false, err
	}

	return trainer, nil
}

func isSelf(ctx context.Context) (bool, error) {
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

func SetupDirectives(config *gql.Config, store *database.Store) {
	config.Directives.IsAdmin = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		admin, err := isAdmin(store, ctx)
		if err != nil {
			return nil, err
		}

		if admin {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsManager = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		manager, err := isManager(store, ctx)
		if err != nil {
			return nil, err
		}

		if manager {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsStaff = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		staff, err := isStaff(store, ctx)
		if err != nil {
			return nil, err
		}

		if staff {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsTrainer = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		trainer, err := isTrainer(store, ctx)
		if err != nil {
			return nil, err
		}

		if trainer {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsSelf = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		self, err := isSelf(ctx)
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
