package models

import "time"

type Hold struct {
	Id         int
	CreatorId  int
	RemoverId  *int
	TargetId   int
	Reason     string
	CreateDate time.Time
	RemoveDate time.Time
}
