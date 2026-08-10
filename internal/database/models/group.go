package models

// Determines how users see the group theyre a part of
type GroupViewPermission int

const (
	// User does not see anyone in the group and does not see that they themselves are in the group
	GroupViewPermission_SeeNone GroupViewPermission = iota
	// User can see that they are in a group but not anyone else
	GroupViewPermission_SeeSelf
	// User can see that they are in a group and see their group mates
	GroupViewPermission_SeeAll
)

// A group is a collection of users
// Groups do not contain permissions or attributes for anything other than how to manage themselves.
// Giving specific permissions is done by attaching a group in the place where the permissions exist rather than adding permissions to the group
// Every group except for the root group has a manager group
// The manager group can view and manage the groups below it
type Group struct {
	Id          int
	ManagerId   int // nullable in database so that the root can be null. However, real groups should not have this be null
	Name        string
	Description string
}

// Stores a users membership in a group
// membership is either direct: A user is manually added to a group with a certain view permission or
// indirect: the user is a direct of a subgroup of the super group with view permissions defined at the subgroup relation
type MembershipToGroup struct {
	UserId         int
	ViewPermission GroupViewPermission
}

type MembershipToUser struct {
	GroupId        int
	ViewPermission GroupViewPermission
}

// Subgroup Links define collections of groups that form a single larger group
// Subgroups behave as if adding a user to a subgroup also adds that user to the supergroup manually with the specified permission
// Removing a user from the subgroup behaves like removing them from the supergroup manually unless they exist in another supergroup
// If a user is in multiple subgroups, the highest permission is shown
type SubgroupLink struct {
	GroupId        int
	SubgroupId     int
	ViewPermission GroupViewPermission
}
