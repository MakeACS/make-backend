package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"make-backend/internal/database/models"
)

type GroupRepository interface {
	GetAdminGroupId(ctx context.Context) (int, error)
	// get group information by ID
	CreateGroup(ctx context.Context, group models.Group) (models.Group, error)
	UpdateGroup(ctx context.Context, group models.Group) (models.Group, error)

	GetGroupById(ctx context.Context, id int) (*models.Group, error)

	// list the combination of all level subgroups
	GetSubgroups(ctx context.Context, groupId int) ([]models.SubgroupLink, error)
	// list the first level subgroups of a given group
	GetDirectSubgroups(ctx context.Context, groupId int) ([]models.SubgroupLink, error)

	AddSubgroupToGroup(ctx context.Context, subgroup, supergroup int, permissions models.GroupViewPermission) error
	IsGroupSubgroup(ctx context.Context, subgroup, supergroup int) (bool, models.GroupViewPermission, error)

	GetGroupMembers(ctx context.Context, groupId int) ([]models.MembershipToGroup, error)
	GetDirectGroupMembers(ctx context.Context, groupId int) ([]models.MembershipToGroup, error)

	IsUserInGroup(ctx context.Context, userId int, groupId int) (bool, models.GroupViewPermission, error)

	AllGroupsUserIsMemberOf(ctx context.Context, userId int) ([]models.MembershipToUser, error)
	IsUserInGroupDirectly(ctx context.Context, userId int, groupId int) (bool, models.GroupViewPermission, error)
	// add a user to a group
	AddUserToGroup(ctx context.Context, userId int, groupId int, perms models.GroupViewPermission) error
	// remove user from a group
	// return true if removed, false if the user was not removed from that group
	RemoveUserFromGroup(ctx context.Context, userId int, groupId int) (bool, error)

	CanUserManageGroup(ctx context.Context, managerUser, managedGroup int) (bool, error)
	CanGroupManageGroup(ctx context.Context, managerGroup, managedGroup int) (bool, error)

	AllGroupsGroupCanManage(ctx context.Context, userId int) ([]int, error)
	AllGroupsVisibleToGroup(ctx context.Context, groupId int) ([]int, error)
	IsGroupVisibleToGroup(ctx context.Context, lookingGroupId int, groupId int) (bool, error)

	AllGroupsUserCanManage(ctx context.Context, userId int) ([]int, error)
	AllGroupsUserVisibleToUser(ctx context.Context, userId int) ([]int, error)
	IsGroupVisibleToUser(ctx context.Context, userId int, groupId int) (bool, error)
}

type GroupRepo struct {
	DB *sql.DB
}

func (g *GroupRepo) IsGroupVisibleToUser(ctx context.Context, userId int, groupId int) (bool, error) {
	// if can manage -> see all
	canManage, err := g.CanUserManageGroup(ctx, userId, groupId)
	if err != nil {
		return false, fmt.Errorf("failed to check if group is visible to user check 1: %w", err)
	}
	if canManage {
		return true, nil
	}
	// TODO check if group is shared to group that user is in

	isInGroup, how, err := g.IsUserInGroup(ctx, userId, groupId)
	if err != nil {
		return false, fmt.Errorf("failed to check if group is visible to user check 2: %w", err)
	}

	if isInGroup {
		switch how {
		case models.GroupViewPermission_SeeAll:
			return true, nil
		case models.GroupViewPermission_SeeNone:
			return false, nil
		case models.GroupViewPermission_SeeSelf:
			return true, nil
		default:
			return false, fmt.Errorf("invalid view permission: %v", how)
		}
	}
	return false, nil
}
func (g *GroupRepo) IsGroupVisibleToGroup(ctx context.Context, lookingGroupId int, groupId int) (bool, error) {
	// if can manage -> see all
	canManage, err := g.CanGroupManageGroup(ctx, lookingGroupId, groupId)
	if err != nil {
		return false, fmt.Errorf("failed to check if group is visible to group 1: %w", err)
	}
	if canManage {
		return true, nil
	}

	isSubgroup, perms, err := g.IsGroupSubgroup(ctx, lookingGroupId, groupId)
	if err != nil {
		return false, fmt.Errorf("failed to check if group is visible to group 3: %w", err)
	}
	if isSubgroup {
		switch perms {

		case models.GroupViewPermission_SeeAll:
			return true, nil
		case models.GroupViewPermission_SeeNone:
			return false, nil
		case models.GroupViewPermission_SeeSelf:
			return true, nil
		default:
			return false, fmt.Errorf("invalid view permission %v", perms)
		}

	}
	return false, nil

}

func (g *GroupRepo) IsGroupSubgroup(ctx context.Context, subgroup int, supergroup int) (bool, models.GroupViewPermission, error) {
	var perms models.GroupViewPermission
	query := `select view_permission from group_subgroups where supergroup_id = $1 and subgroup_id = $2`

	row := g.DB.QueryRowContext(ctx, query, supergroup, subgroup)

	err := row.Scan(&perms)
	if errors.Is(err, sql.ErrNoRows) {
		return false, models.GroupViewPermission_SeeNone, nil
	}

	if err != nil {
		return false, models.GroupViewPermission_SeeNone, err
	}
	return true, perms, nil

}

func (g *GroupRepo) AddSubgroupToGroup(ctx context.Context, subgroupID, supergroupID int, permissions models.GroupViewPermission) error {
	query := `INSERT INTO group_direct_subgroups (group_id, subgroup_id, view_permission) VALUES ($1, $2, $3)`

	_, err := g.DB.ExecContext(ctx, query, supergroupID, subgroupID, permissions)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

func (g *GroupRepo) GetDirectSubgroups(ctx context.Context, groupId int) ([]models.SubgroupLink, error) {
	query := `select subgroup_id, view_permission from group_direct_subgroups where supergroup_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}

	links := []models.SubgroupLink{}
	for rows.Next() {
		var link models.SubgroupLink
		link.GroupId = groupId
		err := rows.Scan(
			&link.SubgroupId,
			&link.ViewPermission,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan direct subgroups: %w", err)
		}
		links = append(links, link)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get direct subgroups: %w", err)
	}

	return links, nil

}

// GetSubgroups implements [GroupRepository].
func (g *GroupRepo) GetSubgroups(ctx context.Context, groupId int) ([]models.SubgroupLink, error) {

	query := `select subgroup_id, view_permission from group_subgroups where supergroup_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}

	links := []models.SubgroupLink{}
	for rows.Next() {
		var link models.SubgroupLink
		link.GroupId = groupId
		err := rows.Scan(
			&link.SubgroupId,
			&link.ViewPermission,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan direct subgroups: %w", err)
		}
		links = append(links, link)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get direct subgroups: %w", err)
	}

	return links, nil

}

// AllGroupsGroupCanManage implements [GroupRepository].
func (g *GroupRepo) AllGroupsGroupCanManage(ctx context.Context, managerId int) ([]int, error) {
	query := `select group_id from group_management where manager_group_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, managerId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}
	defer rows.Close()

	groups := []int{}
	for rows.Next() {
		var group int
		err := rows.Scan(&group)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for groups group can manage: %w", err)
		}
		groups = append(groups, group)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get for groups group can manage: %w", err)
	}
	return groups, nil

}

func (g *GroupRepo) AllGroupsUserCanManage(ctx context.Context, userId int) ([]int, error) {
	query := `
		select gm.group_id 
		from group_management gm 
		left join group_membership m on m.group_id = gm.manager_group_id
		where m.user_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}
	defer rows.Close()

	groups := []int{}
	for rows.Next() {
		var group int
		err := rows.Scan(&group)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for groups user can manage: %w", err)
		}
		groups = append(groups, group)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get for groups user can manage: %w", err)
	}
	return groups, nil
}

func (g *GroupRepo) AllGroupsUserIsMemberOf(ctx context.Context, userId int) ([]models.MembershipToUser, error) {
	query := `select group_id, view_permission from group_membership where user_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}
	defer rows.Close()

	groups := []models.MembershipToUser{}
	for rows.Next() {
		var group models.MembershipToUser
		err := rows.Scan(
			&group.GroupId,
			&group.ViewPermission,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for groups user is member of: %w", err)
		}
		groups = append(groups, group)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get group user is member of: %w", err)
	}

	return groups, nil

}

// AllGroupsUserVisibleToUser implements [GroupRepository].
func (g *GroupRepo) AllGroupsUserVisibleToUser(ctx context.Context, userId int) ([]int, error) {
	panic("unimplemented")
}

// AllGroupsVisibleToGroup implements [GroupRepository].
func (g *GroupRepo) AllGroupsVisibleToGroup(ctx context.Context, groupId int) ([]int, error) {
	panic("unimplemented")
}

// CanGroupManageGroup implements [GroupRepository].
func (g *GroupRepo) CanGroupManageGroup(ctx context.Context, managerGroup int, managedGroup int) (bool, error) {
	var exists bool

	query := "SELECT EXISTS(SELECT 1 FROM group_management WHERE manager_group_id = $1 and group_id = $2)"
	err := g.DB.QueryRow(query, managerGroup, managedGroup).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

func (g *GroupRepo) CanUserManageGroup(ctx context.Context, managerUser int, managedGroup int) (bool, error) {
	var exists bool

	query := `
	SELECT EXISTS(SELECT 1 
			from group_management gm 
			left join group_membership m on m.group_id = gm.manager_group_id
			where m.user_id = $1 and gm.group_id = $2)
	`
	err := g.DB.QueryRow(query, managerUser, managedGroup).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

func (g *GroupRepo) IsUserInGroup(ctx context.Context, userId int, groupId int) (bool, models.GroupViewPermission, error) {
	var perms models.GroupViewPermission

	query := "SELECT view_permission FROM group_membership WHERE user_id = $1 and group_id = $2"

	row := g.DB.QueryRow(query, userId, groupId)

	err := row.Scan(&perms)
	if errors.Is(err, sql.ErrNoRows) {
		return false, models.GroupViewPermission_SeeNone, nil
	}

	if err != nil {
		return false, models.GroupViewPermission_SeeNone, err
	}
	return true, perms, nil
}

func (g *GroupRepo) IsUserInGroupDirectly(ctx context.Context, userId int, groupId int) (bool, models.GroupViewPermission, error) {
	var perms models.GroupViewPermission

	query := "SELECT view_permission FROM group_membership WHERE user_id = $1 and group_id = $2"

	err := g.DB.QueryRow(query, userId, groupId).Scan(&perms)
	if errors.Is(err, sql.ErrNoRows) {
		return false, models.GroupViewPermission_SeeNone, nil
	}
	if err != nil {
		return false, models.GroupViewPermission_SeeNone, err
	}
	return true, perms, nil

}

func (g *GroupRepo) GetAdminGroupId(ctx context.Context) (int, error) {
	var id_result int

	query := `select id from groups where manager_id is null`

	err := g.DB.QueryRowContext(ctx, query).Scan(&id_result)

	if err != nil {
		return 0, fmt.Errorf("failed to get admin group: %w", err)
	}

	return id_result, nil
}

func (g *GroupRepo) CreateGroup(ctx context.Context, group models.Group) (models.Group, error) {
	var result models.Group

	query := `INSERT INTO groups (manager_id, name, description) VALUES ($1, $2, $3) RETURNING id, manager_id, name, description`

	err := g.DB.QueryRowContext(ctx, query, group.ManagerId, group.Name, group.Description).Scan(&result.Id, &result.ManagerId, &result.Name, &result.Description)
	if err != nil {
		return result, fmt.Errorf("failed to create group: %w", err)
	}

	return result, nil
}

func (g *GroupRepo) UpdateGroup(ctx context.Context, group models.Group) (models.Group, error) {
	var result models.Group

	query := `UPDATE groups set manager_id = $1, name = $2, description=$3 RETURNING id, manager_id, name, description where id = $4 RETURNING id, manager_id, name, description`

	err := g.DB.QueryRowContext(ctx, query, group.ManagerId, group.Name, group.Description, group.Id).Scan(&result.Id, &result.ManagerId, &result.Name, &result.Description)
	if err != nil {
		return result, fmt.Errorf("failed to create group: %w", err)
	}

	return result, nil

}

func (g *GroupRepo) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	var result models.Group
	var managerId *int
	query := `SELECT
		id,
		manager_id,
		name,
		description
		FROM groups WHERE id = $1`

	err := g.DB.QueryRowContext(ctx, query, id).Scan(
		&result.Id,
		&managerId,
		&result.Name,
		&result.Description,
	)

	if err != nil {
		return nil, err
	}
	if managerId == nil {
		// was the root
		result.ManagerId = result.Id
	} else {
		result.ManagerId = *managerId
	}

	return &result, nil

}

func (g *GroupRepo) AddUserToGroup(ctx context.Context, userId int, groupId int, permissions models.GroupViewPermission) error {

	query := `INSERT INTO group_direct_membership (group_id, user_id, view_permission) VALUES ($1, $2, $3)`

	_, err := g.DB.ExecContext(ctx, query, groupId, userId, permissions)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

func (g *GroupRepo) GetGroupMembers(ctx context.Context, groupId int) ([]models.MembershipToGroup, error) {
	query := `select user_id, view_permission from group_membership where group_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext for GetGroupMembers: %w", err)
	}

	members := []models.MembershipToGroup{}
	for rows.Next() {
		var member models.MembershipToGroup
		err := rows.Scan(
			&member.UserId,
			&member.ViewPermission,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for group members: %w", err)
		}
		members = append(members, member)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	return members, nil
}

// GetDirectGroupMembers implements [GroupRepository].
func (g *GroupRepo) GetDirectGroupMembers(ctx context.Context, groupId int) ([]models.MembershipToGroup, error) {
	query := `select user_id, view_permission from group_direct_membership where group_id = $1`

	rows, err := g.DB.QueryContext(ctx, query, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}

	members := []models.MembershipToGroup{}
	for rows.Next() {
		var member models.MembershipToGroup
		err := rows.Scan(
			&member.UserId,
			&member.ViewPermission,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for group members: %w", err)
		}
		members = append(members, member)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	return members, nil
}

func (g *GroupRepo) RemoveUserFromGroup(ctx context.Context, userId int, groupId int) (bool, error) {
	query := `DELETE FROM group_direct_memberships where user_id = $1 and groupId = $2`
	result, err := g.DB.ExecContext(ctx, query, groupId, userId)
	if err != nil {
		return false, fmt.Errorf("failed to remove user from group: %w", err)
	}
	num, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check if remove user from group succeeded: %w", err)
	}
	return num > 0, nil
}
