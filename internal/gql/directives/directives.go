package directives

import (
	"context"
	"errors"
	"fmt"
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

	user, err := store.Users.GetUserById(ctx, *user_id)
	if err != nil {
		return false, err
	}

	return user.Admin, nil
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

func extractMakerspaceIdFromFieldNameInMakerspaceType(obj any, makerspaceIdField string) (int, error) {
	// obj represents the parent Go struct (ie. Makerspace)
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Extract sibling field via reflection or direct type assertion
	fieldVal := val.FieldByName(cases.Title(language.AmericanEnglish).String((makerspaceIdField)))
	makerspaceId := 0
	if fieldVal.IsValid() {
		makerspaceId, _ = fieldVal.Interface().(int)
	} else {
		return 0, fmt.Errorf("invalid field to extract. name: %v", makerspaceIdField)
	}
	return makerspaceId, nil

}

func extractMakerspaceIdFromArgNameInMakerspaceType(ctx context.Context, makerspaceIdArg string) (int, error) {
	fc := graphql.GetFieldContext(ctx)

	if idVal, ok := fc.Args[makerspaceIdArg]; ok {
		id := idVal.(int)
		return id, nil
	}
	return 0, fmt.Errorf("couldn't extract id from arg")

}

func extractMakerspaceIdFromArgOrFieldName(ctx context.Context, obj any, makerspaceIdField *string, makerspaceIdArg *string) (int, error) {
	if makerspaceIdArg == nil && makerspaceIdField == nil {
		return 0, errors.New("specify one of makerspaceIdField makerspaceIdArg")
	}
	if makerspaceIdArg != nil && makerspaceIdField != nil {
		return 0, errors.New("specify one of makerspaceIdField makerspaceIdArg not both")
	}

	if makerspaceIdArg != nil {
		return extractMakerspaceIdFromArgNameInMakerspaceType(ctx, *makerspaceIdArg)
	}
	return extractMakerspaceIdFromFieldNameInMakerspaceType(obj, *makerspaceIdField)

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
	config.Directives.IsManagerFor = func(ctx context.Context, obj any, next graphql.Resolver, makerspaceIdField *string, makerspaceIdArg *string) (any, error) {
		user_id := auth.UserIDFromContext(ctx)
		if user_id == nil {
			return false, fmt.Errorf("Access denied")
		}

		makerspaceId, err := extractMakerspaceIdFromArgOrFieldName(ctx, obj, makerspaceIdField, makerspaceIdArg)
		if err != nil {
			return nil, fmt.Errorf("couldn't find makerspace id field: %w", err)
		}

		// TODO add direct getter for in manager group
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

		// TODO add direct getter for in manager group
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

}
