package models

import (
	"time"
)

type User struct {
	Id            int
	Email         string
	FullName      string
	PreferredName string
	Pronouns      string
	JoinDate      time.Time
	SetupComplete bool
	Archived      bool
	Notes         string
	Admin         bool
	ForceArchive  *bool
	CardTag       string
}

func (u User) LogEntity() LogEntity {
	return LogEntity{Id: u.Id, Label: u.PreferredName}
}
