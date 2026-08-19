package resolvers

import (
	"context"
	"fmt"
	"make-backend/internal/auth"
	"make-backend/internal/database"
	"make-backend/internal/database/models"
)

func SliceToPtrSlice[E any](s []E) []*E {
	pointers := make([]*E, 0, len(s))
	for i := range s {
		pointers = append(pointers, &s[i])
	}
	return pointers
}

// helper to check if the use stored in context can manage anonymous group
// returns err if not, nil if can
func CanUserFromContextManageAnonymousGroup(store *database.Store, ctx context.Context, agroupID int) error {
	userId := auth.UserIDFromContext(ctx)
	if userId == nil {
		return auth.ErrNotAuthenticated
	}

	canManage, err := store.Groups.CanUserManageAnonymousGroup(ctx, *userId, agroupID)
	if err != nil {
		return err
	}
	if !canManage {
		return models.ErrNoAnonymousGroupOrWrongPermissions
	}
	return nil
}

func CanUserFromContextSeeGroup(store *database.Store, ctx context.Context, groupID int) (models.GroupViewPermission, error) {
	userId := auth.UserIDFromContext(ctx)
	if userId == nil {
		return models.GroupViewPermission_SeeNone, auth.ErrNotAuthenticated
	}

	visible, how, err := store.Groups.IsUserInGroup(ctx, *userId, groupID)
	if err != nil {
		return models.GroupViewPermission_SeeNone, fmt.Errorf("error checking how user can view group: %w", err)
	}
	visibleByManager, err := store.Groups.CanUserManageGroup(ctx, *userId, groupID)
	if err != nil {
		return models.GroupViewPermission_SeeNone, fmt.Errorf("error checking how user can view group: %w", err)
	}
	visible = visible || visibleByManager
	if !visible {
		return models.GroupViewPermission_SeeNone, models.ErrNoGroupOrWrongPermissions
	}
	if visibleByManager {
		return models.GroupViewPermission_SeeAll, nil
	}
	return how, nil

}
