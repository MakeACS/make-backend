package models

import "time"

type Restriction struct {
	Id           int
	MakerspaceId int
	CreatorId    int
	TargetId     int
	CreateDate   time.Time
	Reason       string
}
