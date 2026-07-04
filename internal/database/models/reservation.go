package models

import "time"

type Reservation struct {
	Id             int
	UserId         *int
	OrganizationId *int
	EquipmentId    int
	Description    string
	StartTime      time.Time
	EndTime        time.Time
	Approved       bool
}
