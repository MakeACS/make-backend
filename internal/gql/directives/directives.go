package directives

import (
	"context"
	"fmt"
	"log/slog"
	"make-backend/internal/auth"
	"make-backend/internal/database"
	"make-backend/internal/gql"
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func isAdmin(store *database.Store, ctx context.Context) (bool, error) {
	user_id := auth.UserIDFromContext(ctx)
	if user_id == nil {
		return false, fmt.Errorf("Access denied")
	}

	adminGroupId, err := store.Groups.GetAdminGroupId(ctx)
	if err != nil {
		return false, err
	}

	isAdmin, _, err := store.Groups.IsUserInGroup(ctx, *user_id, adminGroupId)
	return isAdmin, err
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
		return false, fmt.Errorf("failed to get field context")
	}

	user_id, ok := fc.Args["id"].(int)
	if !ok {
		return false, fmt.Errorf("failed to get id from field context")
	}

	asker_id := auth.UserIDFromContext(ctx)
	if asker_id == nil {
		return false, nil
	}
	slog.Info("isSelf check", "asker", *asker_id, "user", user_id)

	return *asker_id == user_id, nil
}

func extractIdFromFieldNameInType(obj any, idFieldName string) (int, error) {
	// obj represents the parent Go struct (ie. Makerspace, User)
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Extract sibling field via reflection or direct type assertion
	fieldVal := val.FieldByName(cases.Title(language.AmericanEnglish).String((idFieldName)))
	makerspaceId := 0
	if fieldVal.IsValid() {
		makerspaceId, _ = fieldVal.Interface().(int)
	} else {
		return 0, fmt.Errorf("invalid field to extract. name: %v", idFieldName)
	}
	return makerspaceId, nil

}

func extractIdFromArgName(ctx context.Context, idArg string) (int, error) {
	fc := graphql.GetFieldContext(ctx)

	if idVal, ok := fc.Args[idArg]; ok {
		id := idVal.(int)
		return id, nil
	}
	return 0, fmt.Errorf("couldn't extract id from arg")

}

func extractMakerspaceIdFromArgOrFieldName(ctx context.Context, obj any, makerspaceIdField *string, makerspaceIdArg *string) (int, error) {
	return extractIdFromArgOrFieldName("makerspace", ctx, obj, makerspaceIdField, makerspaceIdArg)
}

func extractUserIdFromArgOrFieldName(ctx context.Context, obj any, makerspaceIdField *string, makerspaceIdArg *string) (int, error) {
	return extractIdFromArgOrFieldName("user", ctx, obj, makerspaceIdField, makerspaceIdArg)
}
func extractGroupIdFromArgOrFieldName(ctx context.Context, obj any, makerspaceIdField *string, makerspaceIdArg *string) (int, error) {
	return extractIdFromArgOrFieldName("group", ctx, obj, makerspaceIdField, makerspaceIdArg)
}

func extractIdFromArgOrFieldName(prefix string, ctx context.Context, obj any, idField *string, idArg *string) (int, error) {
	if idArg == nil && idField == nil {
		return 0, fmt.Errorf("specify one of %sIdField or %sIdArg", prefix, prefix)
	}
	if idArg != nil && idField != nil {
		return 0, fmt.Errorf("specify one of %sIdField or %sIdArg not both", prefix, prefix)
	}

	if idArg != nil {
		return extractIdFromArgName(ctx, *idArg)
	}
	return extractIdFromFieldNameInType(obj, *idField)

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

	config.Directives.IsSelf = func(ctx context.Context, obj any, next graphql.Resolver, userIdField, userIdArg *string) (any, error) {
		asker_id := auth.UserIDFromContext(ctx)
		if asker_id == nil {
			// if theres no person signed in, not an error just they cant be self
			return false, nil
		}

		user_id, err := extractUserIdFromArgOrFieldName(ctx, obj, userIdField, userIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find user id field/arg: %w", err)
		}

		if *asker_id == user_id {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}
	config.Directives.IsManagerFor = func(ctx context.Context, obj any, next graphql.Resolver, makerspaceIdField *string, makerspaceIdArg *string) (any, error) {
		user_id := auth.UserIDFromContext(ctx)
		if user_id == nil {
			return false, fmt.Errorf("Access denied")
		}

		makerspaceId, err := extractMakerspaceIdFromArgOrFieldName(ctx, obj, makerspaceIdField, makerspaceIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find makerspace id field: %w", err)
		}

		// TODO: add direct getter for in manager group
		// Or not if permissions go fine grained
		m, err := store.Makerspaces.GetMakerspaceById(ctx, makerspaceId)
		if err != nil {
			return nil, err
		}
		isManager, err := store.Groups.IsUserInAnonymousGroup(ctx, m.ManagementAgroupId, *user_id)
		if err != nil {
			return nil, err

		}

		if isManager {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsStaffFor = func(ctx context.Context, obj any, next graphql.Resolver, makerspaceIdField *string, makerspaceIdArg *string) (any, error) {
		user_id := auth.UserIDFromContext(ctx)
		if user_id == nil {
			return false, fmt.Errorf("Access denied")
		}

		makerspaceId, err := extractMakerspaceIdFromArgOrFieldName(ctx, obj, makerspaceIdField, makerspaceIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find makerspace id field: %w", err)
		}

		// TODO: add direct getter for in manager group
		m, err := store.Makerspaces.GetMakerspaceById(ctx, makerspaceId)
		if err != nil {
			return nil, err
		}
		isStaff, err := store.Groups.IsUserInAnonymousGroup(ctx, m.StaffAgroupId, *user_id)
		if err != nil {
			return nil, err

		}

		if isStaff {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.IsStaffOrManagerFor = func(ctx context.Context, obj any, next graphql.Resolver, makerspaceIdField *string, makerspaceIdArg *string) (any, error) {
		user_id := auth.UserIDFromContext(ctx)
		if user_id == nil {
			return false, fmt.Errorf("Access denied")
		}

		makerspaceId, err := extractMakerspaceIdFromArgOrFieldName(ctx, obj, makerspaceIdField, makerspaceIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find makerspace id field: %w", err)
		}

		// TODO: add direct getter for in manager group
		m, err := store.Makerspaces.GetMakerspaceById(ctx, makerspaceId)
		if err != nil {
			return nil, err
		}
		isStaff, err := store.Groups.IsUserInAnonymousGroup(ctx, m.StaffAgroupId, *user_id)
		if err != nil {
			return nil, err
		}

		if isStaff {
			return next(ctx)
		}

		isManager, err := store.Groups.IsUserInAnonymousGroup(ctx, m.ManagementAgroupId, *user_id)
		if err != nil {
			return nil, err
		}

		if isManager {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.CanAskerManageGroup = func(ctx context.Context, obj any, next graphql.Resolver, groupIdField *string, groupIdArg *string) (any, error) {
		user_id := auth.UserIDFromContext(ctx)
		if user_id == nil {
			return false, fmt.Errorf("Access denied")
		}

		groupId, err := extractGroupIdFromArgOrFieldName(ctx, obj, groupIdField, groupIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find group id field: %w", err)
		}

		canManage, err := store.Groups.CanUserManageGroup(ctx, *user_id, groupId)
		if err != nil {
			return nil, err
		}

		if canManage {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

	config.Directives.CanAskerSeeGroup = func(ctx context.Context, obj any, next graphql.Resolver, groupIdField *string, groupIdArg *string) (any, error) {
		user_id := auth.UserIDFromContext(ctx)
		if user_id == nil {
			return false, fmt.Errorf("Access denied")
		}

		groupId, err := extractGroupIdFromArgOrFieldName(ctx, obj, groupIdField, groupIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find group id field: %w", err)
		}

		canManage, err := store.Groups.IsGroupVisibleToUser(ctx, *user_id, groupId)
		if err != nil {
			return nil, err
		}

		if canManage {
			return next(ctx)
		} else {
			return nil, fmt.Errorf("Unauthorized")
		}
	}

}
