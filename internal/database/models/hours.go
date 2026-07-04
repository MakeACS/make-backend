package models

import "time"

type DefaultHours struct {
	MakerspaceId int
	DayOfWeek    int
	OpenTime     *time.Time
	CloseTime    *time.Time
	Closed       bool
}

type SpecialHours struct {
	MakerspaceId int
	SpecialDate  time.Time
	OpenTime     *time.Time
	CloseTime    *time.Time
	Closed       bool
}
