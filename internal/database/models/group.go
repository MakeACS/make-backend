package models

type GroupViewPermission int

const (
	GroupViewPermission_SeeNone GroupViewPermission = iota
	GroupViewPermission_SeeSelf
	GroupViewPermission_SeeAll
)

type Group struct {
	Id          int
	ManagerId   int // nullable in database so that the root can be null. However, real groups should not be null
	Name        string
	Description string
}

type SubgroupLink struct {
	GroupId        int
	SubgroupId     int
	ViewPermission GroupViewPermission
}

type MembershipToGroup struct {
	UserId         int
	ViewPermission GroupViewPermission
}

type MembershipToUser struct {
	GroupId        int
	ViewPermission GroupViewPermission
}
