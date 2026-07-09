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
		return false, fmt.Errorf("Failed to get target_id from field context")
	}

	user_id, ok := ctx.Value("user_id").(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from session context")
	}

	return target_id == user_id, nil
}

func isManagerFor(store *database.Store, ctx context.Context) (bool, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return false, fmt.Errorf("Failed to get field context")
	}

	makerspace_id, ok := fc.Args["makerspace_id"].(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from field context")
	}

	user_id, ok := ctx.Value("user_id").(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from session context")
	}

	managerFor, err := store.Users.IsManagerFor(ctx, user_id, makerspace_id)
	if err != nil {
		return false, err
	}

	return managerFor, nil
}

func isStaffFor(store *database.Store, ctx context.Context) (bool, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return false, fmt.Errorf("Failed to get field context")
	}

	makerspace_id, ok := fc.Args["makerspace_id"].(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from field context")
	}

	user_id, ok := ctx.Value("user_id").(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from session context")
	}

	staffFor, err := store.Users.IsStaffFor(ctx, user_id, makerspace_id)
	if err != nil {
		return false, err
	}

	return staffFor, nil
}

func isTrainerFor(store *database.Store, ctx context.Context) (bool, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return false, fmt.Errorf("Failed to get field context")
	}

	equipment_id, ok := fc.Args["equipment_id"].(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from field context")
	}

	user_id, ok := ctx.Value("user_id").(int)
	if !ok {
		return false, fmt.Errorf("Failed to get user_id from session context")
	}

	trainerFor, err := store.Users.IsTrainerFor(ctx, user_id, equipment_id)
	if err != nil {
		return false, err
	}

	return trainerFor, nil
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

	config.Directives.IsManagerFor = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		managerFor, err := isManagerFor(store, ctx)
		if err != nil {
			return nil, err
		}

		if managerFor {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsStaffFor = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		staffFor, err := isStaffFor(store, ctx)
		if err != nil {
			return nil, err
		}

		if staffFor {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsTrainerFor = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		trainerFor, err := isTrainerFor(store, ctx)
		if err != nil {
			return nil, err
		}

		if trainerFor {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsStaffOrSelf = func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		staff, err := isStaff(store, ctx)
		if err != nil {
			return nil, err
		}

		self, err := isSelf(ctx)
		if err != nil {
			return nil, err
		}

		if self || staff {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}
}
