package models

import "time"

type MakerspaceHours interface {
	IsMakerspaceHours()
}

type DefaultHours struct {
	MakerspaceId int
	DayOfWeek    int
	OpenTime     *time.Time
	CloseTime    *time.Time
	Closed       bool
}

func (DefaultHours) IsMakerspaceHours() {}

type SpecialHours struct {
	MakerspaceId int
	SpecialDate  time.Time
	OpenTime     *time.Time
	CloseTime    *time.Time
	Closed       bool
}

func (SpecialHours) IsMakerspaceHours() {}
